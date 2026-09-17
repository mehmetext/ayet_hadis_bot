# Ayet Hadis Bot

Bot, Telegram polling ile kullanıcıları yönetir ve planlı her gönderimde QuranEnc veya HadeethEnc API'sinden anlık içerik alır. Ayet/hadis metni, çeviri veya hadis listesi yerelde saklanmaz.

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

Telegram’ı çalıştırmak için `.env` içinde BotFather’dan aldığın token’ı doldur:

```env
TELEGRAM_BOT_TOKEN=bot-token-buraya
```

Token’ı Git’e veya loglara ekleme; `.env` dosyası `.gitignore` içindedir.

`DATA_DIR=data` hem lokal kullanımda proje içindeki `data/bot.db` yolunu, hem Docker içinde kalıcı `bot-data` volume'unu ifade eder. Bu değeri normalde değiştirmen gerekmez.

```sh
docker compose up -d --build
docker compose logs -f
docker compose run --rm ayet-hadis-bot once
```

## CI/CD

`main` branch'ine yapılan her commit, GitHub Actions tarafından test edilip Docker image olarak GHCR'a gönderilir. Pipeline daha sonra SSH ile sunucuya bağlanır, yalnız bu projenin klasöründe `docker compose pull` ve `docker compose up -d` çalıştırır. Sunucuda Docker build yapılmaz ve başka projelerin image'ları temizlenmez.

Sunucuda yalnızca boş bir deployment klasörü oluştur; repository’yi sunucuda clone etmene veya `.env` dosyasını elle yazmana gerek yok. Pipeline her deploy’da güncel `docker-compose.yml` dosyasını ve GitHub’daki ayarlardan üretilen `.env` dosyasını SCP/SSH ile gönderir. GitHub repository ayarlarında şu Actions secret'larını tanımla: `DEPLOY_HOST`, `DEPLOY_USER`, `DEPLOY_PATH`, `DEPLOY_SSH_KEY_B64`, `GHCR_USERNAME`, `GHCR_TOKEN`, `TELEGRAM_BOT_TOKEN`. `GHCR_TOKEN` yalnız `read:packages` yetkisine sahip bir classic PAT olmalı.

`DEPLOY_SSH_KEY_B64`, Actions runner’ında satır sonu bozulmaması için Base64’e çevrilmiş private SSH key olmalıdır. Oluşturmak için lokalinde `base64 -i ~/.ssh/deploy_key | tr -d '\n'` çalıştırıp çıktıyı secret olarak kaydet; private key’in kendisini loglara veya sohbete yapıştırma.

GitHub Actions Variables olarak şu çalışma ayarlarını ekle: `TIMEZONE`, `CONSOLE_LANGUAGE`, `SEND_WINDOW_START`, `SEND_WINDOW_END`, `DAILY_NOTIFICATION_COUNT`, `HTTP_TIMEOUT_SECONDS`, `DATA_DIR`. Pipeline bu değerlerden sunucuda izinleri `0600` olan `.env` dosyasını atomik olarak oluşturur.

API geçici olarak erişilemezse bot eski bir içerik göndermez. Aynı aday için 1, 5, 15 dakika sonra; ardından aktif pencere kapanana kadar 30 dakikada bir yeniden dener.

## Kaynaklar

- [QuranEnc API](https://quranenc.com/nqo/home/api)
- [HadeethEnc API](https://hadeethenc.com)
