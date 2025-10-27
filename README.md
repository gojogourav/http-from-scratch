# HTTP Server From Scratch (in Go)

A lightweight HTTP/1.1 server written **from scratch in Go** — no `net/http` used.  
This project reimplements key parts of HTTP, including request parsing, response writing, and chunked transfer encoding.  

## 🚀 Features

- **Custom HTTP Request Parser**  
  Parses the start line, headers, and message body manually.

- **Custom HTTP Response Writer**  
  Builds status lines, headers, and bodies without relying on Go’s standard library HTTP package.

- **Supports Chunked Transfer Encoding**  
  Implements real-time streaming of data in chunks using:


- **Supports Video handling**  
  Implements real-time streaming of data in chunks using:

