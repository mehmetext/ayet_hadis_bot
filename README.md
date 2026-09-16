# Ayet Hadis Bot

Telegram bağlantısı henüz eklenmemiştir. Bot, planlı her gönderimde QuranEnc veya HadeethEnc API'sinden anlık içerik alır ve konsola yazar. Ayet/hadis metni, çeviri veya hadis listesi yerelde saklanmaz.

SQLite yalnızca gönderim sırası, tekrar engeli ve başarısız denemelerin durumunu `/data/bot.db` içinde tutar. Desteklenen diller kodda sabittir: `ara`, `eng`, `tur`, `deu`.

## Komutlar

```sh
go run ./cmd/ayet-hadis-bot once # Anlık API'den tek içerik alır.
go run ./cmd/ayet-hadis-bot run  # Zaman penceresinde çalışır.
```

`sync` komutu yoktur; ilk kurulumda arşiv indirilmez.

## Ayarlar

```env
TIMEZONE=Europe/Istanbul
SEND_WINDOW_START=06:30
SEND_WINDOW_END=22:30
DAILY_NOTIFICATION_COUNT=4
DATA_DIR=data
```

Docker Compose, `DATA_DIR` için kalıcı `bot-data` volume'unu `/data` altında bağlar:

```sh
docker compose up -d --build
docker compose logs -f
docker compose run --rm ayet-hadis-bot once
```

API geçici olarak erişilemezse bot eski bir içerik göndermez. Aynı aday için 1, 5, 15 dakika sonra; ardından aktif pencere kapanana kadar 30 dakikada bir yeniden dener.

## Kaynaklar

- [QuranEnc API](https://quranenc.com/nqo/home/api)
- [HadeethEnc API](https://hadeethenc.com)
