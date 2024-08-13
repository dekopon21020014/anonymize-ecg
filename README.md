# 心電図データ匿名化アプリ
## 使用言語
![Next.js](https://img.shields.io/badge/Next.js-000000?style=for-the-badge&logo=nextdotjs&logoColor=white)
![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)

## 概要
- 心電図データ(mwfおよびxml)に含まれる個人情報を削除し匿名化するwebアプリケーションです

## 環境
- Node.js 18.17.1(docker)
- Go 1.22.1(docker)

## インストール手順
```bash
git clone https://github.com/shikidalab/anonymize-ecg.git
cd anonymize-ecg
mv .env.sample .env # DOWNLOAD_DIRに，csvをダウンロードしたいパス(ローカル)を指定する
mv ./back/.env.sample ./back/.env # DSNと，csvのダウンロード先のパス(コンテナ内)を指定する
docker compose run -w /app --rm front npm install 
docker compose up