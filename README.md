# Ayet Hadis Bot

Telegram bağlantısı henüz eklenmemiştir. Bu sürüm, QuranEnc ve HadeethEnc'den ilk eşitlemede aldığı içeriği kalıcı SQLite arşivinde tutar ve konsola yazar.

Desteklenen teslim dilleri kod içinde sabittir: `ara`, `eng`, `tur`, `deu`. Ortam değişkenleri içerik kaynağını veya dili değiştirmez.

## Komutlar

```sh
go run ./cmd/ayet-hadis-bot sync  # Eksik yerel arşivi indirir.
go run ./cmd/ayet-hadis-bot once  # SQLite'tan bir ayet/hadis seçer.
go run ./cmd/ayet-hadis-bot run   # Eşitleyip pencere içindeki slotlarda çalışır.
```

Operasyonel ayarlar:

```env
TIMEZONE=Europe/Istanbul
SEND_WINDOW_START=06:30
SEND_WINDOW_END=22:30
DAILY_NOTIFICATION_COUNT=4
DATA_DIR=/data
```

Docker'da SQLite dosyası `bot-data` volume'unda `/data/bot.db` olarak kalır:

```sh
docker compose up --build
docker compose run --rm ayet-hadis-bot sync
docker compose run --rm ayet-hadis-bot once
```

Bu ilk sürüm Telegram bağlantısı olmadan, seçilen bir ayet ve hadisi konsola yazdırır. Gönderim aralığı `CONTENT_INTERVAL` ile şimdiden yapılandırılabilir; Telegram scheduler'ı sonraki kapsamdır.

## Çalıştırma

Yerelde:

```bash
cp .env.example .env
go run ./cmd/ayet-hadis-bot
```

Docker ile:

```bash
cp .env.example .env
docker compose up -d --build
docker compose logs -f
```

`QURAN_VERSE_NUMBER`, `QURAN_EDITION`, `HADITH_NUMBER` ve `HADITH_EDITION` değerleri `.env` içinden değiştirilebilir. `CONTENT_INTERVAL` şu an gelecekteki gönderim scheduler'ı için saklanır; örneğin `3h`, `4h30m` veya `10h`.

## Veri kaynakları

- Ayetler: [AlQuran.cloud API](https://alquran.cloud/api). Anahtarsızdır; kullanımda IP başına saniyelik yumuşak rate limit uygulanır. Bu uygulama seyrek istek yaptığı için API sınırlarıyla uyumludur.
- Hadisler: [fawazahmed0/hadith-api](https://github.com/fawazahmed0/hadith-api). Statik JSON CDN uçları, kitap ve hadis numarası bazında ayrıntı sağlar; proje README'si ücretsiz ve rate limit'siz olduğunu belirtir.

Ücretsiz dış servislerin sürekliliği garanti edilmediğinden, Telegram aşamasında alınan içerikleri yerel cache'e kaydetmek ve tekrar göndermemek için gönderim geçmişi tutmak gerekir.
