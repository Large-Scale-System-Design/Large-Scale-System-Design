# Example
## Send email
```bash
curl -X POST http://localhost:8080/send \
     -H "Content-Type: application/json" \
     -d '{
           "channel": "email",
           "recipient": "your-email@gmail.com",
           "properties": {
             "subject": "Strategy Pattern Test",
             "body": "It works dynamically!"
           }
        }'
```
