# Reply Bot
予め呼び掛けと返答の組を設定しておくと、メンション付きでいずれかの呼び掛けが投稿されたときに対応する文で返信を返すActivityPub実装です。  
C3クリエイタソン Snowカップにて作成を開始した作品です。  

## Requirements
ビルドにはGoが必要です。
想定されるGoのバージョンは[go.mod](./go.mod)を参照してください。  

## Usage
以下のいずれの場合も、別途リバースプロキシなどを適切に設定することを推奨します。  
### Non-Docker
1. リポジトリ全体をコピーし、リポジトリルートに移動します。  
   ```bash
   git clone https://github.com/Bokume2/reply_bot.git
   cd reply_bot
   ```
1. .envに各設定値を記述します。  
   ```bash
   cp .env.sample .env
   vim .env  # edit .env
   ```
1. reply_dialogues.yamlに、Botが行う掛け合いを記述します。  
   ```bash
   vim reply_dialogues.yaml
   ```
1. Goをインストールします。  
1. 初回起動前に初期データを挿入します。  
   ```bash
   go run scripts/seeds.go
   ```
1. サーバーを起動します。ポート3000でlistenを開始します。  
   ```bash
   go run cmd/main.go
   ```

### Docker
1. リポジトリ全体をコピーし、リポジトリルートに移動します。  
   ```bash
   git clone https://github.com/Bokume2/reply_bot.git
   cd reply_bot
   ```
1. .envに各設定値を記述します。  
   ```bash
   cp .env.sample .env
   vim .env  # edit .env
   ```
1. reply_dialogues.yamlに、Botが行う掛け合いを記述します。  
   ```bash
   vim reply_dialogues.yaml
   ```
1. Dockerをインストールします。  
1. コンテナをビルドします。  
   ```bash
   docker compose build
   ```
1. 初回起動前に初期データを挿入します。  
   ```bash
   docker compose run --rm app /bin/seeds
   ```
1. サーバーを起動します。ポート3000でlistenを開始します。  
   ```bash
   docker compose up -d
   ```

## Development
Visual Studio CodeのDev Container拡張機能に対応しています。VSCodeを開発コンテナで起動後、各種設定ファイルを記述してから開発コンテナのターミナルで`air`を実行するとホットリロード付きの開発サーバーが起動します。  
あるいは、.devcontiner/compose.yamlを使用して直接Docker Composeで起動することもできます。この場合、コンテナの起動と同時に自動で`air`が実行されます。  

## Samples
samplesディレクトリ内のファイルはreply_dialogues.yamlの設定例です。  
設定時の参考になれば幸いです。  

## License
**ファイル内に明確に例外表示されたものを除いて**、ソースコードは[Unlicenseライセンス](https://unlicense.org)で配布されています。  
詳しくは[ライセンス表示](./UNLICENSE)または[https://unlicense.org](https://unlicense.org)を参照して下さい。  

## Contact
不具合や機能提案などのご連絡は[Twitter(現X)](https://x.com/boku_renraku)やその他までお気軽にお声がけください。  
