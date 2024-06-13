# Gcalendar
Untuk mendapatkan token pertama kali,juga jika token habis buka ini dan jalankan di lokal:
```go
tok, err = GetTokenFromWeb(config)
if err != nil {
    return nil, err
}
err = SaveToken(db, tok)
if err != nil {
    return nil, err
}
```