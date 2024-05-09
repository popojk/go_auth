DB_NAME=coffee_shop_finder_user_db
CONTAINER=$DB_NAME

# create database
echo -e "\n### Creating database $DB_NAME ###"
docker exec -i $CONTAINER psql -U postgres -c "CREATE DATABASE $DB_NAME ENCODING UTF8 LC_COLLATE='C.UTF-8' LC_CTYPE='C.UTF-8';"