All of the sensetive data that should not be shown in a real app is hidden, but for testing purposes
.env should look like this\
{\
SERVER_PORT=8085\
DB_HOST=db\
DB_PORT=5432\
DB_USER=postgres\
DB_PASSWORD=postgres\
DB_NAME=subscriptions\
DB_SSLMODE=disable\
LOG_LEVEL=info\
}\
P.S. docker-compose.yml also has to have corresponding fields updated


