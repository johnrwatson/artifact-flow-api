# REST API for Artifact Records

This is a REST API implemented in Golang that allows users to manage artifact records. The API supports basic CRUD operations (Create, Read, Update, Delete) for managing artifact data.

## Requirements

- Go 1.20 or above
- MongoDB

## Installation Locally

1. Clone the repository:

```
git clone https://github.com/artifact-flow/artifact-flow-api.git
```

2. Change to the project directory:

```
cd artifact-flow-api
```

3. Install dependencies:

```
go mod download 
```

4. Configure MongoDB database locally:
   
```
docker compose -f ./local-development/docker-compose-mongo.yml up -d
```

5. Run the application:

```
go run main.go
```

The API server should now be running on `http://localhost:8000`.

## API Endpoints

### Create an Artifact Record

**Request**

```
POST /artifacts
```

Create a new artifact record.

**Request Body**

```json
{
  "name": "Artifact Name",
  "description": "Artifact Description",
  "category": "Artifact Category"
}
```

**Response**

```json
{
  "id": "123",
  "name": "Artifact Name",
  "description": "Artifact Description",
  "category": "Artifact Category"
}
```

### Get All Artifact Records

**Request**

```
GET /artifacts
```

Retrieve all artifact records.

**Response**

```json
[
  {
    "id": "123",
    "name": "Artifact Name",
    "description": "Artifact Description",
    "category": "Artifact Category"
  },
  {
    "id": "456",
    "name": "Another Artifact",
    "description": "Another Description",
    "category": "Another Category"
  }
]
```

### Get a Specific Artifact Record

**Request**

```
GET /artifacts/{id}
```

Retrieve a specific artifact record by its ID.

**Response**

```json
{
  "id": "123",
  "name": "Artifact Name",
  "description": "Artifact Description",
  "category": "Artifact Category"
}
```

### Update an Artifact Record

**Request**

```
PUT /artifacts/{id}
```

Update an existing artifact record.

**Request Body**

```json
{
  "name": "Updated Artifact Name",
  "description": "Updated Artifact Description",
  "category": "Updated Artifact Category"
}
```

**Response**

```json
{
  "id": "123",
  "name": "Updated Artifact Name",
  "description": "Updated Artifact Description",
  "category": "Updated Artifact Category"
}
```

### Delete an Artifact Record

**Request**

```
DELETE /artifacts/{id}
```

Delete an existing artifact record.

**Response**

```
Artifact record deleted successfully.
```

## Contributing

Contributions are welcome! If you find any issues or want to add new features, please submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).