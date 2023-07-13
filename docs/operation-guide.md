# SAKURA Distributed MariaDB Controller(Sakura-DBC) オペレーションガイド

## 概要

Sakura-DBCの動作確認方法、操作方法について説明します。

## Sakura-DBCの状態遷移

Sakura-DBCは、内部的に以下の4つの状態を持ち、状況に応じて状態遷移を行います。

- fault状態
  - DBサーバとしての機能を停止している状態
  - 以下の場合にこの状態に遷移する
    - 起動直後
    - ネットワーク分断を検知した場合
    - デュアルprimary状態を検知した場合
    - 実行したコマンドがエラーとなった場合
    - 自身がfault状態で、対向のDBサーバがcandidateやreplica状態の場合
    - (対向のDBサーバがprimaryに遷移するのを待つためfault状態にとどまる)
- candidate状態
  - primaryに遷移しようとしている状態
  - 対向DBサーバと同時にprimary(デュアルprimary)とならないよう、一時的にこの状態に遷移する
  - 以下の場合にこの状態に遷移する
    - 自身がfault、もしくはreplica状態において、対向DBサーバがfaultになった場合
- primary状態
  - primaryデータベースとして動作している状態
  - クライアントからのDB接続を受け付け、DBリクエストを処理できる状態
  - MariaDBはread_only=0に設定し、3306番ポートへの接続を許可するnftablesルールを設定する
  - 以下の場合にこの状態に遷移する
    - 自身がcandidateになった後、対向DBサーバがcandidate、もしくはprimaryでない場合
- replica状態
  - replicaデータベースとして動作している状態
  - primaryに対してレプリケーションを張り、更新データを受信している状態
  - MariaDBはread_only=1に設定し、3306番ポートへの接続を拒否するnftablesルールを設定する
  - 以下の場合にこの状態に遷移する
    - 自身がfaultの状態で、対向DBサーバがprimaryの場合

## BGP経路の属性

Sakura-DBCは、BGP経路のCommunity属性として表現することで、他のノードに自身の状態を広告します。

| 状態      | BGP Community |
| --------- | ------------- |
| fault     | 65001:1       |
| candidate | 65001:2       |
| primary   | 65001:3       |
| replica   | 65001:4       |
| anchor    | 65001:10      |

## ログの確認方法

Sakura-DBCは、状態遷移や、それに伴い実行したコマンドなどをログ出力します。ログを確認するにはjournalctlコマンドを利用します。

```
journalctl -u sakura-controller -e
```

## 現在の状態の確認方法

Sakura-DBCの現在の状態を確認するには、curlコマンドなどで以下のエンドポイントをHTTPリクエストします。

```
curl http://127.0.0.1:54545/status
{"state":"replica"}
```

## BGP経路の確認方法

BGPピア状態の確認方法を以下に示します。

```
[root@test-db1 ~]# vtysh -c 'show ip bgp summary'

IPv4 Unicast Summary (VRF default):
BGP router identifier xx.xx.xx.xx, local AS number 65001 vrf-id 0
BGP table version 4
RIB entries 5, using 960 bytes of memory
Peers 2, using 1449 KiB of memory

Neighbor        V         AS   MsgRcvd   MsgSent   TblVer  InQ OutQ  Up/Down State/PfxRcd   PfxSnt Desc
xx.xx.xx.xx     4      65001    228427    228427        0    0    0 01w0d22h            2        3 N/A
xx.xx.xx.xx     4      65001    228427    228427        0    0    0 01w0d22h            2        3 N/A

Total number of neighbors 2
```

BGP経路情報の確認方法を以下に示します。

```
[root@test-db1 ~]# vtysh -c 'show ip bgp'
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

[root@test-db1 ~]# vtysh -c 'show ip bgp community-list primary'
<snip>

    Network          Next Hop            Metric LocPrf Weight Path
 * ixx.xx.xx.xx/32   xx.xx.xx.xx              0    100      0 i
 *>i                 xx.xx.xx.xx              0    100      0 i

Displayed  1 routes and 5 total paths
```

## Sakura-DBCの起動

Sakura-DBCを起動するには以下のようにコマンドを入力します。

```
[root@test-db1 ~]# systemctl start sakura-controller
[root@test-db1 ~]# systemctl status sakura-controller
● sakura-controller.service - Database Controller
   Loaded: loaded (/etc/systemd/system/sakura-controller.service; enabled; vendor preset: disabled)
   Active: active (running) since Thu 2023-07-13 16:56:21 JST; 4s ago
 Main PID: 1391344 (sakura-controll)
    Tasks: 9 (limit: 24876)
   Memory: 5.5M
   CGroup: /system.slice/sakura-controller.service
           └─1391344 /root/distributed-mariadb-controller/bin/sakura-controller --log-level info --db-repilica-password-filepath /root/.db-replica-password
<snip>
```

## Sakura-DBCの停止

Sakura-DBCを停止するには以下のようにコマンドを入力します。

```
[root@test-db1 ~]# systemctl stop sakura-controller
[root@test-db1 ~]# systemctl status sakura-controller
● sakura-controller.service - Database Controller
   Loaded: loaded (/etc/systemd/system/sakura-controller.service; enabled; vendor preset: disabled)
   Active: inactive (dead) since Thu 2023-07-13 16:55:35 JST; 7s ago
  Process: 694 ExecStart=/root/distributed-mariadb-controller/bin/sakura-controller --log-level info --db-repilica-password-filepath /root/.db-replica-password (code=exited, status=0/SUCCESS)
 Main PID: 694 (code=exited, status=0/SUCCESS)
<snip>
```

## ログレベルの変更方法

[クイックスタートガイド](quick-start-guide.md)の手順では、通常の運用において推奨されるinfoログレベルにて設定するようになっています。
ログレベルを変更するには、以下のようにします。

```
vi /etc/systemd/system/sakura-controller.service

! infoになっている部分を変更します
ExecStart = /root/distributed-mariadb-controller/bin/sakura-controller --log-level info --db-repilica-password-filepath /root/.db-replica-password
```

systemdに反映し、Sakura-DBCを再起動します。

```
systemctl daemon-reload
systemctl restart sakura-controller
```

指定可能なログレベルと、各ログレベルにおける出力の指針は以下の通りです。

- debug
  - 開発中や異常発生時に、デバッグを行う必要がある場合のログレベル
  - ソースコードの実行パスを必要十分に追える粒度での外部インターフェイスの振る舞い
  - 内部状態、意思決定、全ての実行コマンド(作用/副作用のないものも含め)を出力
- info
  - 主要な状態遷移、主要な意思決定、作用/副作用の発生する実行コマンドを出力
- warning
  - サービス停止には至らないが、例外ケースが発生するなど1営業日程度でオペレータの確認を要するログ
- error
  - サービス停止に至る、もしくはその可能性が高い異常ケースで、直ちにオペレータの確認を要するログ
