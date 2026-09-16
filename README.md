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

Tüm çalışma ayarlarının tek kaynağı proje kökündeki `.env` dosyasıdır. Bir kez oluştur:

```sh
cp .env.example .env
```

Ardından saat aralığı ve bildirim sayısını yalnız `.env` içinde değiştir. Uygulama lokal çalışırken bu dosyayı kendisi okur; Docker Compose da aynı dosyayı konteynere verir. Eksik bir değer varsa uygulama hangi değerin eksik olduğunu söyleyerek başlatmayı durdurur.

`CONSOLE_LANGUAGE` konsol akışının dilini (`ara`, `eng`, `tur` veya `deu`) belirler. Telegram kullanıcı tercihleri eklenene kadar gönderimler bu tek dil için yapılır.

`DATA_DIR=data` hem lokal kullanımda proje içindeki `data/bot.db` yolunu, hem Docker içinde kalıcı `bot-data` volume'unu ifade eder. Bu değeri normalde değiştirmen gerekmez.

```sh
docker compose up -d --build
docker compose logs -f
docker compose run --rm ayet-hadis-bot once
```

API geçici olarak erişilemezse bot eski bir içerik göndermez. Aynı aday için 1, 5, 15 dakika sonra; ardından aktif pencere kapanana kadar 30 dakikada bir yeniden dener.

## Kaynaklar

- [QuranEnc API](https://quranenc.com/nqo/home/api)
- [HadeethEnc API](https://hadeethenc.com)
