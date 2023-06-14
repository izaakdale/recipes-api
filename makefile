run:
	CLIENT_ID="xxxxxxxxxxxxxxx" \
	CLIENT_SECRET="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
	AUTH0_DOMAIN=dev-np0m9-q3.us.auth0.com \
	AUTH0_API_IDENTIFIER="https://api.recipes.io" \
	MONGO_URI="mongodb://admin:password@localhost:27017/test?authSource=admin" \
	MONGO_DATABASE=demo \
	go run .

swag:
	swagger generate spec -o ./swagger.yaml