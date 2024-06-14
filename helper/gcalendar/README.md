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

gmail

```go
// Define the email details
    to := "recipient@example.com"
    subject := "Test Email with Attachment"
    body := "This is a test email with attachment."
    attachmentPaths := []string{"path/to/attachment1.pdf", "path/to/attachment2.jpg"}

    // Send the email
    err = gcalendar.SendEmailWithAttachment(db, to, subject, body, attachmentPaths)
    if err != nil {
        log.Fatalf("Error sending email: %v", err)
    }
```

```go
// Define the email details
 to := "recipient@example.com"
 subject := "Test Email"
 body := "This is a test email."

 // Send the email
 err = gcalendar.SendEmail(db, to, subject, body)
 if err != nil {
  log.Fatalf("Error sending email: %v", err)
 }

```

blogger

```go
 // Define the blog details
 blogID := "your-blog-id" // Ganti dengan ID blog Anda
 title := "Test Post"
 content := "This is a test post."
 content := `
  <h1>This is a test post</h1>
  <p>This is a paragraph with <strong>bold</strong> text and <em>italic</em> text.</p>
  <ul>
   <li>Item 1</li>
   <li>Item 2</li>
   <li>Item 3</li>
  </ul>
 `

 // Post to Blogger
 post, err := gcalendar.PostToBlogger(db, blogID, title, content)
 if err != nil {
  log.Fatalf("Error posting to Blogger: %v", err)
 }

 blogID := "your-blog-id"
 postID := "your-post-id"

 err = gcalendar.DeletePostFromBlogger(db, blogID, postID)
 if err != nil {
  log.Fatalf("Failed to delete post: %v", err)
 }

```

drive

```go
fileID := "your-file-id"
 newTitle := "Duplicated File"

 duplicatedFile, err := gcalendar.DuplicateFileInDrive(db, fileID, newTitle)
 if err != nil {
  log.Fatalf("Failed to duplicate file: %v", err)
 }
```

docs

```go
docID := "your-doc-id"
 replacements := map[string]string{
  "oldText1": "newText1",
  "oldText2": "newText2",
 }

 err = gcalendar.ReplaceStringsInDoc(db, docID, replacements)
 if err != nil {
  log.Fatalf("Failed to replace strings in document: %v", err)
 }
```

pdf

```go
docID := "your-doc-id"
 outputFileName := "output.pdf"

 fileID, err := gcalendar.GeneratePDF(db, docID, outputFileName)
 if err != nil {
  log.Fatalf("Failed to generate PDF: %v", err)
 }

```
