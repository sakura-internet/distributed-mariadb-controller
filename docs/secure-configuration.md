# SAKURA Distributed MariaDB Controller(Sakura-DBC) セキュアコンフィグレーションガイド

## 概要

Sakura-DBCをセキュアに動作させるために、nftablesを用いたパケットフィルタの設定例を紹介します。

クイックスタートガイドの手順と併せて実施することを強く推奨します。

## 各ポートへのアクセス制限

Sakura-DBC、および関連ソフトウェアが利用するポート番号は以下の通りです。
これらのポートに対してアクセス元IPアドレス制限をかけることを推奨します。

| ポート番号 | 利用するデーモン | 用途 | 制限するアクセス元 |
| ---------- | ---------------- | ---- | ------------------ |
| 22 | sshd | SSHリモートアクセス | SSHログインが必要なアクセス元 |
| 179 | bgpd(FRRouting) | BGPピア | BGPピアを張る対向のデータベースサーバやアンカーサーバ |
| 3306 | mariadbd | DBアクセス(MySQLプロトコル) | ※ sakura-controllerにて自動設定 |
| 50505 | sakura-controller | Prometheus exporter | Metrics取集元であるPrometheusサーバ |
| 54545 | sakura-controller | GSLBヘルスチェック | GSLBヘルスチェック元 |

※ Sakura-DBCは、nftablesを用いて3306番ポートのアクセス許可/拒否ルールを設定します

## nftablesの設定例

nftablesの設定例を以下に示します。RHEL8系のOSでは、 `/etc/sysconfig/nftables.conf` に作成してください。

nftables.conf
```
table ip filter {
    set ssh_allow_src {
        type ipv4_addr
        flags interval
        elements = { 192.0.2.0/24, 203.0.113.0/24 }   ← SSHログイン元IPアドレスを記載
    }
    chain ssh {
        type filter hook input priority filter; policy accept;
        iifname "eth0" ip saddr @ssh_allow_src tcp dport 22 accept
        iifname "eth0" tcp dport 22 drop
    }

    set prometheus_src {
        type ipv4_addr
        flags interval
        elements = { 192.0.2.0/32, 203.0.113.0/32 }   ← PrometheusサーバのIPアドレス(/32単位)を記載
    }
    chain prometheus {
        type filter hook input priority filter; policy accept;
        iifname "eth0" ip saddr @prometheus_src tcp dport 50505 accept
        iifname "eth0" tcp dport 50505 drop
    }

    set gslb_healthcheck_src {
        type ipv4_addr
        flags interval
        elements = { 192.0.2.0/24, 203.0.113.0/24 }   ← GSLBヘルスチェック元のIPアドレスを記載
    }
    chain gslb_health_check {
        type filter hook input priority filter; policy accept;
        iifname "eth0" ip saddr @gslb_healthcheck_src tcp dport 54545 accept
        iifname "eth0" tcp dport 54545 drop
    }

    set bgp_allow_src {
        type ipv4_addr
        flags interval
        elements = { 192.0.2.0/32, 203.0.113.0/32 }   ← 対向のデータベースサーバ、アンカーサーバのIPアドレス(/32単位)を記載
    }
    chain bgp {
        type filter hook input priority filter; policy accept;
        iifname "eth0" ip saddr @bgp_allow_src tcp dport 179 accept
        iifname "eth0" tcp dport 179 drop
    }
}
```

ルールを反映させます。

```
systemctl enable nftables
systemctl restart nftables
```

ルールが反映されたか確認します。

```
nft list ruleset
```
