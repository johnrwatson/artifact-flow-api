# REST API for Artifact Records

This is a REST API implemented in Golang that allows users to manage artifacts and their qualification for environments.

The API supports basic CRUD operations (Create, Read, Update, Delete) for managing artifact & qualification data.

The follows

## Requirements

- Go 1.20 or above
- MongoDB

##  Prerequisites
Before running the application, make sure to set the following environment variables in either a docker compose file or k8s manifest:

Database Connection:
`DB_USERNAME`: The username for the MongoDB database.
`DB_PASSWORD`: The password for the MongoDB database.
`DB_CONNECTION_STRING`: The connection string for the MongoDB database. (defaults to mongodb://:@localhost:27017)

Disable Authentication:
`OPEN_ENDPOINTS`: For local development disable all authentication by setting this to true.

If Authentication is Enabled:
`OAUTH_CLIENT_ID`: The OAuth client ID for authentication.
`OAUTH_CLIENT_SECRET`: The OAuth client secret for authentication.
`OAUTH_REDIRECT_URL`: The redirect URL for OAuth authentication.
`OAUTH_SESSION_SECRET`: The secret key for session management.

## Installation Locally

1. Clone the repository:

```
git clone https://github.com/artifact-flow/artifact-flow-api.git
```

2. Change to the project directory:

```
cd artifact-flow-api
```

3. Run the application:

```
make deploy
```

The API server should now be running on `http://localhost:80`.

## API Endpoints

The following API endpoints are available in the application:

### Artifacts

- **Create Artifact**
  - URL: `/artifacts`
  - Method: `POST`
  - Handler Function: `artifacts.CreateArtifact`
  - Authentication: `Bearer` (If authentication enabled)

*Request Body:*
```json
{
  "name": "Artifact Name",                  # Optional
  "description": "Artifact Description",    # Optional
  "artifactType": "Artifact Type",          # Optional
  "artifactFamily": "",                     # Optional
  "artifactMetadata": {                     # Nested/Extensible map of values and key pairs
    "key1": "value1",
    "subkey1": {
      "subkey1key1": "subkey1key1value"
    }
  }
}
```

- **Get Artifacts**
  - URL: `/artifacts`
  - Method: `GET`
  - Handler Function: `artifacts.GetArtifacts`
  - Authentication: `Bearer` (If authentication enabled)

- **Search Artifacts**
  - URL: `/artifacts/search`
  - Method: `POST`
  - Handler Function: `artifacts.SearchArtifacts`
  - Authentication: `Bearer` (If authentication enabled)

*Request Body:*
```json
{
  "searchKey": "artifactFamily",         # Required (the key from the artifact record to search by)
  "searchValue": "example-family",       # Required (the value of the key from the artifact to search by)
  "searchVerb": "equal",                 # Required (one of contains/equal)
}
```

- **Get Artifact by ID**
  - URL: `/artifacts/{id}` # `Where id is the ID of the artifact requested`
  - Method: `GET`
  - Handler Function: `artifacts.GetArtifact`
  - Authentication: `Bearer` (If authentication enabled)

- **Update Artifact**
  - URL: `/artifacts/{id}` # `Where id is the ID of the artifact requested`
  - Method: `PUT`
  - Handler Function: `artifacts.UpdateArtifact`
  - Authentication: `Bearer` (If authentication enabled)

*Request Body:*
```json
{
  "name": "Artifact Name",                  # Optional
  "description": "Artifact Description",    # Optional
  "artifactType": "Artifact Type",          # Optional
  "artifactFamily": "",                     # Optional
  "artifactMetadata": {                     # Nested/Extensible map of values and key pairs
    "key1": "value1",
    "subkey1": {
      "subkey1key1": "subkey1key1value"
    }
  }
}
```

- **Delete Artifact**
  - URL: `/artifacts/{id}` # `Where id is the ID of the artifact requested`
  - Method: `DELETE`
  - Handler Function: `artifacts.DeleteArtifact`
  - Authentication: `Bearer` (If authentication enabled)

### Validation Rules

- **Create Rule**
  - URL: `/validation/rules`
  - Method: `POST`
  - Handler Function: `validation.CreateRule`
  - Authentication: `Bearer` (If authentication enabled)

*Request Body:*
```json
{
    "name": "Sample Validation Rule 2",             # Optional
    "description": "Example Validation Rule 2",     # Optional
    "ruleFamily": "code",                           # Optional
    "ruleLimits": [                                 # Required: Extensible map of limits. 
        {
        "type": "equal",                            # Required: one of equal/min/max/set 
        "value": "melons"                           # Required: value to compare against
        }
    ],
    "ruleKey": "artifactFamily"                     # Required: the key to apply the rule to
}
```

- **Get Rules**
  - URL: `/validation/rules`
  - Method: `GET`
  - Handler Function: `validation.GetRules`
  - Authentication: `Bearer` (If authentication enabled)

- **Get Rule by ID**
  - URL: `/validation/rules/{id}` # `Where id is the ID of the validation rule requested`
  - Method: `GET`
  - Handler Function: `validation.GetRule`
  - Authentication: `Bearer` (If authentication enabled)

- **Search Rules**
  - URL: `/validation/rules/search`
  - Method: `POST`
  - Handler Function: `validation.SearchRules`
  - Authentication: `Bearer` (If authentication enabled)

```json
{
  "searchKey": "ruleFamily",             # Required (the key from the rule record to search by)
  "searchValue": "example-family",       # Required (the value of the key from the rule to search by)
  "searchVerb": "equal",                 # Required (one of contains/equal)
}
```

- **Update Rule**
  - URL: `/validation/rules/{id}` # `Where id is the ID of the validation rule requested`
  - Method: `PUT`
  - Handler Function: `validation.UpdateRule`
  - Authentication: `Bearer` (If authentication enabled)

*Request Body:*
```json
{
    "name": "Sample Validation Rule 2",             # Optional
    "description": "Example Validation Rule 2",     # Optional
    "ruleFamily": "code",                           # Optional
    "ruleLimits": [                                 # Required: Extensible map of limits. 
        {
        "type": "equal",                            # Required: one of equal/min/max/set 
        "value": "melons"                           # Required: value to compare against
        }
    ],
    "ruleKey": "artifactFamily"                     # Required: the key to apply the rule to
}
```

- **Delete Rule**
  - URL: `/validation/rules/{id}` # `Where id is the ID of the validation rule requested`
  - Method: `DELETE`
  - Handler Function: `validation.DeleteRule`
  - Authentication: `Bearer` (If authentication enabled)

### Validation Rule Mappings

- **Create Rule Mapping**
  - URL: `/validation/mappings`
  - Method: `POST`
  - Handler Function: `validation.CreateRuleMapping`
  - Authentication: `Bearer` (If authentication enabled)

*Request Body:*
```json
{
    "name": "Sample Validation Mapping",            # Optional
    "ruleId": "649ff5ad32ae554426073b9b",           # Required: The Rule ID to apply
    "enforced": true,                               # Optional: Currently ignored by codebase
    "environments": {                               # Optional: environments to apply the rule to
        "dev": true,                                # true/false currently assumed to be true for all environments
        "preprod": false
    }
}
```

- **Get Rule Mappings**
  - URL: `/validation/mappings`
  - Method: `GET`
  - Handler Function: `validation.GetRuleMappings`
  - Authentication: `Bearer` (If authentication enabled)

- **Get Rule Mapping by ID**
  - URL: `/validation/mappings/{id}` # `Where id is the ID of the validation mapping requested`
  - Method: `GET`
  - Handler Function: `validation.GetRuleMapping`
  - Authentication: `Bearer` (If authentication enabled)

- **Search Rule Mappings**
  - URL: `/validation/mappings/search`
  - Method: `POST`
  - Handler Function: `validation.SearchRuleMappings`
  - Authentication: `Bearer` (If authentication enabled)

```json
{
  "searchKey": "environments.dev",       # Required (the key from the rule mapping record to search by)
  "searchValue": "true",                 # Required (the value of the key from the rule mapping to search by)
  "searchVerb": "equal",                 # Required (one of contains/equal)
}
```

- **Update Rule Mapping**
  - URL: `/validation/mappings/{id}` # `Where id is the ID of the validation mapping requested`
  - Method: `PUT`
  - Handler Function: `validation.UpdateRuleMapping`
  - Authentication: `Bearer` (If authentication enabled)

*Request Body:*
```json
{
    "name": "Sample Validation Mapping",            # Optional
    "ruleId": "649ff5ad32ae554426073b9b",           # Required: The Rule ID to apply
    "enforced": true,                               # Optional: Currently ignored by codebase
    "environments": {                               # Optional: environments to apply the rule to
        "dev": true,                                # true/false currently assumed to be true for all environments
        "preprod": false
    }
}
```

- **Delete Rule Mapping**
  - URL: `/validation/mappings/{id}` # `Where id is the ID of the validation mapping requested`
  - Method: `DELETE`
  - Handler Function: `validation.DeleteRuleMapping`
  - Authentication: `Bearer` (If authentication enabled)

### Validation of Artifacts

- **Validate Artifact**
  - Description: `Validates whether an artifact meets all rules applied to the environment`
  - URL: `/validation/artifacts`
  - Method: `POST`
  - Handler Function: `validation.ValidateArtifact`
  - Authentication: `Bearer` (If authentication enabled)

```json
{
  "artifactId": "64a02de5e84e540c589e3ff9",     # Required
  "environment": "dev"                          # Required
}
```

### Authentication and Supporting Handlers

- **Health Check**
  - URL: `/health`
  - Method: `GET`
  - Handler Function: `supporting.Health`

- **Login**
  - URL: `/auth/login`
  - Method: `GET`
  - Handler Function: `auth.LoginHandler`

- **OAuth Callback**
  - URL: `/auth/callback`
  - Method: `GET`
  - Handler Function: `auth.CallbackHandler`

- **Generate Static API Key for Artifact-Flow (api-keys generated are not yet accepted)**
  - URL: `/auth/apikey`
  - Method: `GET`
  - Handler Function: `auth.ApiKeyHandler`
  - Authentication: `Bearer` (If authentication enabled)

## Error Handling

If the prerequisites for database and OAuth provider initialization fail, an error message will be printed, and the application will exit with a status code of `1`.

```go
Error: Prerequisites unable to be initialized:
 - Database Available: <database availability>
 - OAuth Provider: <OAuth provider initialization>
```

If there is an error connecting to the MongoDB database, an error message will be logged.

```go
Error: Unable to connect to the MongoDB database.
```

## Contributing

Contributions are welcome! If you find any issues or want to add new features, please submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).