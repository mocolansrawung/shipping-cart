# Shipping Cart Service API

Shipping Cart Service API is a service to manipulating product, cart, and order related to ecommerce implementation. This project is served as a final project of my Trainee program and basically consists of two microservices along with this service.

## Getting Started

These instructions will guide you on how to run the Shipping Cart Service on your local machine.

### Prerequisites

You need to have Go installed on your machine. This project is developed using Go 1.18.1. While it may work with other versions, it's advisable to use the version that the project was developed with.

### Running the Project

1. Download or clone this repository to your local machine.

2. Navigate to the root directory of the project.

3. Start the application by running the command: 
    ```
    make run
    ```
    This command will start the program. If a database file does not already exist, the program will create one. It will then start serving on the port specified in the `.env` file. By default, it is configured to use port `8081`. 

You should now have the service running locally on your machine. You can interact with it using any HTTP client, or by using the provided Swagger UI.

## API Documentation

The API documentation for this project is available through Swagger UI. Once the project is running, you can access the Swagger UI at the following address:

http://localhost:8081/swagger/index.html

In the Swagger UI, you will be able to see all available endpoints, their expected input and output, and you can even test the endpoints directly in the UI.