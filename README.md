# Ayet & Hadis Botu

QuranEnc ve HadeethEnc kaynaklarından ayet ve hadisleri belirlenen saatlerde Telegram abonelerine gönderen, Go ve SQLite tabanlı açık kaynak bir bot.

Telegram botu: [@hadis_ayet_bot](https://t.me/hadis_ayet_bot)

## Özellikler

- Telegram polling ile çalışır; webhook veya reverse proxy gerektirmez.
- Kullanıcı `/start` ile kayıt olur ve dilini seçer.
- Arapça, İngilizce, Türkçe ve Almanca desteklenir.
- Ayet ve hadis gönderimleri dönüşümlü ilerler.
- Aynı dildeki aboneler ortak içerik akışını paylaşır.
- Her gün belirlenen saat penceresinde eşit aralıklı bildirim gönderir.
- İçerik metinleri yerelde cache’lenmez; SQLite yalnızca sıra, tekrar geçmişi ve kullanıcı durumunu tutar.
- API veya Telegram geçici olarak erişilemezse retry/pending mekanizması kullanılır.
- Docker image’ı GitHub Container Registry’de tutulur ve `main` commit’leri otomatik deploy edilir.

## Telegram komutları

```text
/start     Abone ol ve bildirim dilini seç
/stop      Bildirimleri durdur
/language  Bildirim dilini değiştir
/status    Dil ve abonelik durumunu göster
/help      Komut listesini göster
```

`/start` sonrasında kullanıcıya botun çalışma saatleri ve ayet/hadis akışı açıklanır. Dil seçilmeden abonelik başlatılmaz. Abone olunduğu anda içerik gönderilmez; kullanıcı bir sonraki planlı bildirime dahil olur.

## Çalışma modeli

Bot her bildirim slotunda:

1. Aktif aboneliği olan dilleri SQLite’tan bulur.
2. O dil için sıradaki içerik türünü (`ayet` veya `hadis`) okur.
3. QuranEnc veya HadeethEnc API’sinden tek içeriği alır.
4. Aynı dildeki aktif abonelere gönderir.
5. Başarılı teslimden sonra sıra ve tekrar geçmişini transaction içinde günceller.

İçerik metni, çeviri veya hadis listesi diske yazılmaz. Uzun Telegram mesajları 4096 karakter sınırını aşarsa eksiltilmeden ardışık parçalara bölünür.

## Gereksinimler

- Go 1.26 veya üzeri
- Docker ve Docker Compose
- Telegram BotFather token’ı

## Lokal çalıştırma

Önce ayar dosyasını oluştur:

```sh
cp .env.example .env
```

`.env` içinde Telegram token’ını ve çalışma ayarlarını doldur. Ardından:

```sh
go run ./cmd/ayet-hadis-bot once
```

Bu komut API’den tek bir içerik alıp konsola yazdırır; Telegram’a mesaj göndermez.

Scheduler ve Telegram polling’i birlikte çalıştırmak için:

```sh
go run ./cmd/ayet-hadis-bot run
```

## Docker ile çalıştırma

```sh
cp .env.example .env
# .env dosyasını düzenle
docker compose up -d --build
docker compose logs -f
```

SQLite verisi `bot-data` isimli kalıcı Docker volume’unda tutulur. Konteyner yeniden başlatıldığında kullanıcılar, sıra ve tekrar geçmişi korunur.

## Yapılandırma

Çalışma ayarlarının tek kaynağı `.env` dosyasıdır:

```env
TIMEZONE=Europe/Istanbul
CONSOLE_LANGUAGE=ara
SEND_WINDOW_START=06:30
SEND_WINDOW_END=22:30
DAILY_NOTIFICATION_COUNT=4
HTTP_TIMEOUT_SECONDS=30
DATA_DIR=data
TELEGRAM_BOT_TOKEN=
```

`06:30–22:30` ve `DAILY_NOTIFICATION_COUNT=4` için slotlar başlangıç ve bitiş dahil edilerek hesaplanır: `06:30`, `11:50`, `17:10`, `22:30`.

`.env` dosyası ve Telegram token’ı kesinlikle commit edilmemelidir. `.env` `.gitignore` ve `.dockerignore` ile dışarıda tutulur.

## CI/CD ve sunucu kurulumu

`main` branch’ine yapılan her commit [GitHub Actions workflow’u](.github/workflows/deploy.yml) tarafından test edilir, Docker image olarak GHCR’a gönderilir ve SSH üzerinden sunucuya deploy edilir.

Sunucuda repository clone edilmez ve Docker build yapılmaz. Pipeline yalnızca güncel `docker-compose.yml` dosyasını, GitHub ayarlarından oluşturulan `.env` dosyasını ve GHCR image’ını kullanır.

GitHub Actions Secrets:

```text
DEPLOY_HOST
DEPLOY_USER
DEPLOY_PATH
DEPLOY_SSH_KEY_B64
GHCR_USERNAME
GHCR_TOKEN
TELEGRAM_BOT_TOKEN
```

GitHub Actions Variables:

```text
TIMEZONE
CONSOLE_LANGUAGE
SEND_WINDOW_START
SEND_WINDOW_END
DAILY_NOTIFICATION_COUNT
HTTP_TIMEOUT_SECONDS
DATA_DIR
```

`DEPLOY_SSH_KEY_B64`, satır sonu sorunlarını önlemek için Base64’e çevrilmiş private SSH key olmalıdır. `GHCR_TOKEN` private image pull etmek için en az `read:packages` yetkisine sahip classic PAT olmalıdır.

## Kaynaklar

- [QuranEnc API](https://quranenc.com/nqo/home/api)
- [HadeethEnc](https://hadeethenc.com)

## Katkıda bulunma

Katkı sağlamak isteyenler şu akışı kullanabilir:

1. Repository’yi fork’la.
2. Yeni bir branch oluştur:

   ```sh
   git checkout -b feature/aciklama
   ```

3. Değişiklikleri yap ve testleri çalıştır:

   ```sh
   go test ./...
   ```

4. Commit oluşturup fork’ına push et.
5. `main` branch’ine Pull Request aç.

Pull Request açıklamasında yapılan değişikliği, test sonuçlarını ve varsa davranış değişikliklerini belirt. Yeni özelliklerde mevcut API, SQLite tekrar engeli, Telegram komutları ve Docker çalışma modelinin bozulmadığını doğrula.

## Lisans

Bu repository için lisans henüz belirtilmemiştir. Katkı göndermeden önce repository sahibinin lisans kararını kontrol et.
