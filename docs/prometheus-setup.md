# Prometheusセットアップガイド

## 概要

Prometheusを用いてSakura-DBCをモニタリングする設定例を紹介します。

本ドキュメントでは、異常検知時にSlack通知する手順を記載しています

## 事前準備

あらかじめ、以下を準備してください。

- Prometheusを動作させるサーバ
  - 本ドキュメントでは、Rocky Linux 8を例に説明します
- Slack通知用のチャンネル
- Slack通知用のIncoming WebHook URL

## Prometheusのインストール

Prometheus本体とAlertmanagerをインストールします。

/etc/yum.repos.d/prometheus.repo ファイルを作成します。

```
[prometheus]
name=prometheus
baseurl=https://packagecloud.io/prometheus-rpm/release/el/$releasever/$basearch
repo_gpgcheck=1
enabled=1
gpgkey=https://packagecloud.io/prometheus-rpm/release/gpgkey
       https://raw.githubusercontent.com/lest/prometheus-rpm/master/RPM-GPG-KEY-prometheus-rpm
gpgcheck=1
metadata_expire=300
```

インストールします。

```
update-crypto-policies --set DEFAULT:SHA1
yum -y install prometheus
yum -y install alertmanager
```

## Prometheusの設定

/etc/prometheus/prometheus.yml ファイルを作成します。

```
global:
  scrape_interval: 15s
  evaluation_interval: 15s
rule_files:
  - /etc/prometheus/rules/db-controller.yaml
scrape_configs:
- job_name: prometheus
  static_configs:
  - targets:
    - xx.xx.xx.xx:50505 ← DBサーバのIPアドレス
    - xx.xx.xx.xx:50505 ← DBサーバのIPアドレス
alerting:
  alertmanagers:
  - static_configs:
    - targets:
      - 127.0.0.1:9093
```

/etc/prometheus/rules/db-controller.yaml ファイルを作成します。

```
groups:
- name: db-controller-alert-group
  rules:
  - alert: DBControllerStateCandidate
    expr: edb_db_controller_state{state="candidate"} == 1
    for: 2m
    annotations:
      summary: "DBController State is going down to candidaite"
      description: "{{ $labels.instance }} of job {{ $labels.job }} is now candidate state"
  - alert: DBControllerStateFault
    expr: edb_db_controller_state{state="fault"} == 1
    for: 2m
    annotations:
      summary: "DBController State is going down to fault"
      description: "{{ $labels.instance }} of job {{ $labels.job }} is now fault state"
  - alert: InstanceDown
    expr: up == 0
    for: 2m
    labels:
      severity: page
    annotations:
      summary: "Instance {{ $labels.instance }} down"
      description: "{{ $labels.instance }} of job {{ $labels.job }} has been down for more than 2 minutes."
```

/etc/prometheus/alertmanager.yml ファイルを作成します。

```
global:
  slack_api_url: "https://hooks.slack.com/services/xxx/yyy/zzz" ← SlackのIncoming WebHook URLを記入
route:
  receiver: 'slack-notifications'
  group_wait: 10s
  group_interval: 10s
receivers:
- name: 'slack-notifications'
  slack_configs:
  - channel: 'alert-test' ← 通知先Slackチャンネルを指定
    send_resolved: true
    title: "{{ range .Alerts }}{{ .Annotations.summary }}\n{{ end }}"
    text: "{{ range .Alerts }}{{ .Annotations.description }}\n{{ end }}"
inhibit_rules:
```

PrometheusとAlertmanagerを起動します。

```
systemctl enable prometheus
systemctl start prometheus

systemctl enable alertmanager
systemctl start alertmanager
```

## 通知のトリガーと発報テスト

以下の場合に通知を行うような設定となっていますので、通知が届くかどうか確認してください。

- sakura-controllerデーモンがダウンしている場合(サーバがダウンしている状態も含む)
- primary, replica以外の状態の場合
