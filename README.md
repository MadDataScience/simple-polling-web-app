# simple-polling-web-app

A simple polling web application

```bash
cd cmd
docker build -t application-tag .
docker run -it --rm -p 5051:5050 application-tag
```

spent 1 hour (12/03 22:00-23:00) thinking about it, sketching data model, etc.
spent 1 hour (12/04 13:00-14:00) setting up repo, docker, etc.

To Do:

- Not updating questions on returning
- Authentication - https://blog.usejournal.com/authentication-in-golang-c0677bcce1a8
- Handle errors
- make code modular and dry
