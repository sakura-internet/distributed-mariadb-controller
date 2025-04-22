# SAKURA Distributed MariaDB Controller(Sakura-DBC) オペレーションガイド

## 概要

Sakura-DBCの動作確認方法、操作方法について説明します。

## Sakura-DBCの状態遷移

Sakura-DBCは、内部的に以下の4つの状態を持ち、状況に応じて状態遷移を行います。

- fault状態
  - DBサーバとしての機能を停止している状態
  - 以下の場合にこの状態に遷移します
    - 起動直後
    - ネットワーク分断を検知した場合
    - デュアルprimary状態を検知した場合
    - 実行したコマンドがエラーとなった場合
    - 自身がfault状態で、対向のDBサーバがcandidateやreplica状態の場合(対向のDBサーバがprimaryに遷移するのを待つためfault状態にとどまる)
- candidate状態
  - primaryに遷移しようとしている状態
  - 対向DBサーバと同時にprimary(デュアルprimary)とならないよう、一時的にこの状態に遷移します
  - 以下の場合にこの状態に遷移します
    - 自身がfault、もしくはreplica状態において、対向DBサーバがfaultになった場合
- primary状態
  - primaryデータベースとして動作している状態
  - クライアントからのDB接続を受け付け、DBリクエストを処理できる状態です
  - MariaDBに対し、read_onlyフラグを0に設定し、3306番ポートへの接続を許可するnftablesルールを設定します
  - 以下の場合にこの状態に遷移します
    - 自身がcandidateになった後、対向DBサーバがcandidate、もしくはprimaryでない場合
- replica状態
  - replicaデータベースとして動作している状態
  - primaryに対してレプリケーションを張り、更新データを受信している状態です
  - MariaDBに対し、read_onlyフラグを1に設定し、3306番ポートへの接続を拒否するnftablesルールを設定します
  - 以下の場合にこの状態に遷移します
    - 自身がfaultの状態で、対向DBサーバがprimaryの場合

## BGP経路の属性

Sakura-DBCは、BGP経路のCommunity属性として表現することで、他のノードに自身の状態を広告します。
データベースサーバでは、デーモンとして起動するdb-controllerにgobgpが組み込まれており、BGPピアを確立します。

| 状態      | BGP Community |
| --------- | ------------- |
| fault     | 65000:1       |
| candidate | 65000:2       |
| primary   | 65000:3       |
| replica   | 65000:4       |
| anchor    | 65000:10      |

## Sakura-DBCの起動

Sakura-DBCを起動するには以下のようにコマンドを入力します。

```
# systemctl start db-controller
# systemctl status db-controller
● db-controller.service - Database Controller
   Loaded: loaded (/etc/systemd/system/db-controller.service; enabled; vendor preset: disabled)
   Active: active (running) since Thu 2023-07-13 16:56:21 JST; 4s ago
 Main PID: 1391344 (db-controller)
    Tasks: 9 (limit: 24876)
   Memory: 5.5M
   CGroup: /system.slice/db-controller.service
           └─1391344 /root/distributed-mariadb-controller/bin/db-controller --log-level info --db-replica-password-filepath /root/.db-replica-password
<snip>
```

## Sakura-DBCの停止

Sakura-DBCを停止するには以下のようにコマンドを入力します。

```
# systemctl stop db-controller
# systemctl status db-controller
● sakura-controller.service - Database Controller
   Loaded: loaded (/etc/systemd/system/db-controller.service; enabled; vendor preset: disabled)
   Active: inactive (dead) since Thu 2023-07-13 16:55:35 JST; 7s ago
  Process: 694 ExecStart=/root/distributed-mariadb-controller/bin/db-controller --log-level info --db-replica-password-filepath /root/.db-replica-password (code=exited, status=0/SUCCESS)
 Main PID: 694 (code=exited, status=0/SUCCESS)
<snip>
```

## ログの確認方法

Sakura-DBCは、状態遷移や、それに伴い実行したコマンドなどをログ出力します。ログを確認するにはjournalctlコマンドを利用します。

```
# journalctl -u db-controller -e
```

## 現在の内部状態の確認方法

Sakura-DBCの現在の状態を確認するには、curlコマンドなどで以下のエンドポイントをHTTPリクエストします。

```
# curl http://127.0.0.1:54545/status
{"state":"replica"}
```

## GSLB応答状況の確認方法

Sakura-DBCがGSLBに対し、どのようにレスポンスを行っているか確認するには、curlコマンドなどで以下のエンドポイントをHTTPリクエストします。

```
! primaryの場合(200 OKが返る)
# curl -v http://127.0.0.1:54545/healthcheck
* Connected to 127.0.0.1 (127.0.0.1) port 54545 (#0)
> GET /healthcheck HTTP/1.1
> Host: 127.0.0.1:54545
>
< HTTP/1.1 200 OK
< Date: Wed, 19 Jul 2023 05:19:10 GMT
< Content-Length: 0

! primary以外の場合(503 Service Unavailableが返る)
# curl -v http://127.0.0.1:54545/healthcheck
* Connected to 127.0.0.1 (127.0.0.1) port 54545 (#0)
> GET /healthcheck HTTP/1.1
> Host: 127.0.0.1:54545
>
< HTTP/1.1 503 Service Unavailable
< Date: Wed, 19 Jul 2023 05:19:00 GMT
< Content-Length: 0
```

## BGP経路の確認方法

### アンカーサーバ

BGPピアの状態を確認するには、以下のようにvtyshコマンドを用います。

```
# vtysh -c 'show ip bgp summary'

IPv4 Unicast Summary (VRF default):
BGP router identifier xx.xx.xx.xx, local AS number 65001 vrf-id 0
BGP table version 4
RIB entries 5, using 960 bytes of memory
Peers 2, using 1449 KiB of memory

Neighbor        V         AS   MsgRcvd   MsgSent   TblVer  InQ OutQ  Up/Down State/PfxRcd   PfxSnt Desc
xx.xx.xx.xx     4      65001    228427    228427        0    0    0 01w0d22h            2        3 N/A
xx.xx.xx.xx     4      65002    228427    228427        0    0    0 01w0d22h            2        3 N/A

Total number of neighbors 2
```

BGP経路情報を確認するには、以下のようにvtyshコマンドを用います。

```
# vtysh -c 'show ip bgp'
BGP table version is 4, local router ID is xx.xx.xx.xx, vrf id 0
Default local pref 100, local AS 65001
Status codes:  s suppressed, d damped, h history, * valid, > best, = multipath,
               i internal, r RIB-failure, S Stale, R Removed
Nexthop codes: @NNN nexthop's vrf id, < announce-nh-self
Origin codes:  i - IGP, e - EGP, ? - incomplete
RPKI validation codes: V valid, I invalid, N Not found

    Network          Next Hop            Metric LocPrf Weight Path
 *> xx.xx.xx.xx/32
                    0.0.0.0                  0         32768 i
 *>ixx.xx.xx.xx/32
                     xx.xx.xx.xx           0    100      0 i
 * i                 xx.xx.xx.xx           0    100      0 i
 * ixx.xx.xx.xx/32   xx.xx.xx.xx            0    100      0 i
 *>i                 xx.xx.xx.xx            0    100      0 i

Displayed  3 routes and 5 total paths

# vtysh -c 'show ip bgp community-list primary'
<snip>

    Network          Next Hop            Metric LocPrf Weight Path
 * ixx.xx.xx.xx/32   xx.xx.xx.xx              0    100      0 i
 *>i                 xx.xx.xx.xx              0    100      0 i

Displayed  1 routes and 5 total paths
```

### DBサーバ

BGPピアの状態を確認するには、以下のようにgobgpコマンドを用います。
db-controllerがgobgpのgRPCポート(50051)を待ち受けており、そこから情報が取得されます。

```
# gobgp neighbor
Peer              AS  Up/Down State       |#Received  Accepted
xx.xx.xx.xx 65001 00:30:02 Establ      |        1         1
xx.xx.xx.xx 65002 00:31:05 Establ      |        3         2
```

BGP経路情報を確認するには、以下のようにgobgpコマンドを用います。

```
# gobgp global rib
   Network              Next Hop             AS_PATH              Age        Attrs
*> xx.xx.xx.xx/32     xx.xx.xx.xx        65003                00:00:42   [{Origin: i} {Med: 0} {Communities: 65000:10}]
*  xx.xx.xx.xx/32     xx.xx.xx.xx        65001 65003          00:00:38   [{Origin: i} {Communities: 65000:10}]
*> xx.xx.xx.xx/32     xx.xx.xx.xx                             00:00:35   [{Origin: i} {Communities: 65000:4}]
*> xx.xx.xx.xx/32     xx.xx.xx.xx              65001          00:00:38   [{Origin: i} {Communities: 65000:3}]
*  xx.xx.xx.xx/32     xx.xx.xx.xx        65003 65001          00:00:39   [{Origin: i} {Communities: 65000:3}]
```

## ログレベルの変更方法

[クイックスタートガイド](quick-start-guide.md)の手順では、通常の運用において推奨されるinfoログレベルにて設定するようになっています。
もし、ログレベルを変更するには、以下のようにします。

```
# vi /etc/systemd/system/db-controller.service

! ログレベルを変更します(debugに変更する場合)
ExecStart = /root/distributed-mariadb-controller/bin/db-controller --log-level debug ...
```

systemdに反映し、db-controllerを再起動します。

```
# systemctl daemon-reload
# systemctl restart db-controller
```

指定可能なログレベルと、各レベルにおいて出力されるログの基準は以下の通りです。

- debug
  - 開発中や異常発生時にデバッグを行う場合に使用するログレベル
  - 外部インターフェイスの振る舞い、内部状態の変化、意思決定、全ての実行コマンド(作用/副作用のないものも含め)を出力
- info
  - 通常の運用において必要十分な情報を出力するログレベル
  - 主要な状態遷移や意思決定、作用/副作用の発生する実行コマンドを出力
- warning
  - サービス停止には至らないが、例外ケースが発生した場合など、オペレータの確認を要するログ
- error
  - サービス停止に至る可能性が高い異常ケースが発生した場合など、直ちにオペレータの確認を要するログ
