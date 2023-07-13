# SAKURA Distributed MariaDB Controller(Sakura-DBC) オペレーションガイド

## 概要

Sakura-DBCの動作確認方法、操作方法について説明します。

## Sakura-DBCの状態遷移

Sakura-DBCは、内部的に以下の4つの状態を持ち、状況に応じて状態遷移を行います。

- fault状態
  - DBサーバとしての機能を停止している状態
  - 起動直後やネットワーク分断を検知した場合、デュアルprimaryを検知した場合、その他エラーが発生した場合この状態に遷移する
  - また、対向のDBサーバがcandidateやreplica状態の場合、それがprimaryに遷移するのを待つために、自身はfault状態にとどまる
- candidate状態
  - primaryに遷移しようとしている状態
  - 対向DBサーバと同時にprimary(デュアルprimary)とならないよう、一時的にこの状態に遷移する
- primary状態
  - primaryデータベースとして動作している状態
  - クライアントからのDB接続を受け付け、DBリクエストを処理できる状態
  - MariaDBはread_only=0に設定し、3306番ポートへの接続を許可するnftablesルールを設定する
- replica状態
  - replicaデータベースとして動作している状態
  - primaryに対してレプリケーションを張り、更新データを受信している状態
  - MariaDBはread_only=1に設定し、3306番ポートへの接続を拒否するnftablesルールを設定する

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
