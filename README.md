# SAKURA Distributed MariaDB Controller(Sakura-DBC)

## 概要

SAKURA Distributed MariaDB Controller(以降Sakura-DBCと表記)は、マルチAZ(Availability Zone)環境でMariaDBのフェイルオーバを制御するツールです。

地理的に離れた拠点間でデータを保護し、災害などによる耐障害性を高めることでDR(Disaster Recovery)を実現することを目標としています。

BGPをコントロールプレーンとして使用し、データ不整合の原因となるスプリットブレインの発生を防止する設計となっています。

参考情報として、下記の記事でアーキテクチャについて解説しています。

https://knowledge.sakura.ad.jp/35102/

## インストール

- [クイックスタートガイド](docs/quick-start-guide.md)
  - Sakura-DBCを最小限動作させる手順についてはこちらを参照してください
- [セキュアコンフィグレーションガイド](docs/secure-configuration.md)
  - Sakura-DBCをセキュアに動作させるための設定例はこちらを参照してください
  - クイックスタートガイドと併せて実施することを強く推奨します
- [オペレーションガイド](docs/operation-guide.md)
  - Sakura-DBCの動作確認方法、操作方法についてはこちらを参照してください
- [Prometheusセットアップガイド](docs/prometheus-setup.md)
  - Prometheusを用いてSakura-DBCのモニタリングを行う設定例はこちらを参照してください

## ライセンス

SAKURA Distributed MariaDB Controller Copyright (C) 2023 [The Sakura-DBC Authors](AUTHORS).

This project is published under [Apache 2.0 License](LICENSE.txt).
