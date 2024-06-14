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
## Cara pakai
Isi model saja
```go
simpleEvent := SimpleEvent{
        Summary:     "Google I/O 2024",
        Location:    "800 Howard St., San Francisco, CA 94103",
        Description: "A chance to hear more about Google's developer products.",
        Date:        "2024-06-14",
        TimeStart:   "09:00:00",
        TimeEnd:     "17:00:00",
        Attendees:   []string{"awangga@gmail.com", "awangga@ulbi.ac.id"},
    }
```