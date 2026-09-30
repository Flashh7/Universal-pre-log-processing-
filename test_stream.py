import socket
import time

HOST = "localhost"
PORT = 5050

with socket.create_connection((HOST, PORT)) as s:

    # Authenticate ONCE
    s.sendall(b"payment-service abc123\n")

    # Now send multiple logs over the SAME connection
    logs = [
        '127.0.0.1 - frank [30/Sep/2026:10:15:32 +0530] "GET /index.html HTTP/1.1" 200 1024',
        '192.168.1.20 - alice [30/Sep/2026:10:16:05 +0530] "POST /api/login HTTP/1.1" 200 512',
        '{"timestamp":"2026-09-30T10:30:00+05:30","method":"GET","path":"/api/users","protocol":"HTTP/1.1","status":200}',
        '{"timestamp":"2026-09-30T10:31:00+05:30","action":"login","user":"alice","result":"success"}',
        '<34>Sep 30 10:40:00 server01 sshd[1234]: Failed password for bob from 10.0.0.5',
        'hello this is not a supported log format',
    ]

    for log in logs:
        s.sendall((log + "\n").encode())
        time.sleep(1)

    time.sleep(2)
