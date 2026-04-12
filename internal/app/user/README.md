Создать пользователя:

curl -X POST http://localhost:8080/sign-up \
  -H "Content-Type: application/json" \
  -d '{"email":"user4","password":"password4"}'


Авторизовать пользователя: 

curl -X POST http://localhost:8080/sign-in \
  -H "Content-Type: application/json" \
  -d '{"email":"user4","password":"password4"}'