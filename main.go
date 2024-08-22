package main

import (
    "context"
    "fmt"
    "log"
    "os"

	"github.com/gofiber/fiber/v2"
    "github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
    ID        primitive.ObjectID    `json:"_id,omitempty" bson:"_id,omitempty"`
    Completed bool   `json:"completed"`
    Body      string `json:"body"`
}

var collection *mongo.Collection

func main() {
    fmt.Println("Hello World")
    err := godotenv.Load(".env")
    if err != nil {
        log.Fatal("Error loading .env file", err)
    }

    MONGODB_URI := os.Getenv("MONGODB_URI")
    clientOptions := options.Client().ApplyURI(MONGODB_URI)
    client, err := mongo.Connect(context.Background(), clientOptions)
    
    if err != nil {
        log.Fatal(err)
    }

	defer client.Disconnect(context.Background())

    err = client.Ping(context.Background(), nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Connected to MongoDB!")

    collection = client.Database("go_db").Collection("todos")

	app := fiber.New()

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodo)
	app.Patch("/api/todos/:id", updateTodo)
	app.Delete("/api/todos/:id", deleteTodo)

    PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "5001"
	}

	log.Fatal(app.Listen(":" + PORT))
}

	func getTodos(c *fiber.Ctx) error {
		var todos []Todo
		cursor, err := collection.Find(context.Background(), bson.M{})

		if err != nil {
			return err
		}

		defer cursor.Close(context.Background())

		for cursor.Next(context.Background()) {
			var todo Todo
			if err := cursor.Decode(&todo); err != nil {
				return err
			}
			todos = append(todos, todo)
		}
		return c.JSON(todos)
	}

	func createTodo(c *fiber.Ctx) error {
		todo := new(Todo)

		if err := c.BodyParser(todo); err != nil {
			return err
		}

		if todo.Body == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Body is required"})
		}

		insertResult, err := collection.InsertOne(context.Background(), todo)
		if err != nil {
			return err
		}

		todo.ID = insertResult.InsertedID.(primitive.ObjectID)

		return c.Status(201).JSON(todo)
	}

	func updateTodo(c *fiber.Ctx) error {
		id := c.Params("id")
		ObjectID, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
		}

		filter := bson.M{"_id": ObjectID}
		update := bson.M{"$set": bson.M{"completed": true}}

		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return err
		}

		return c.Status(200).JSON(fiber.Map{"success": true})
	}

	func deleteTodo(c *fiber.Ctx) error {
		id := c.Params("id")
		objectID, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
		}

		filter := bson.M{"_id": objectID}
		_, err = collection.DeleteOne(context.Background(), filter)

		if err != nil {
			return err
		}

		return c.Status(200).JSON(fiber.Map{"success": true})
	}

    // // Get a todo
    // app.Get("/api/todos", func(c *fiber.Ctx) error {
    // 	return c.Status(200).JSON(fiber.Map{"msg": "Hello, World"})
    // })
    // // Create a todo
    // app.Post("/api/todos", func(c *fiber.Ctx) error {
    // 	todo := &Todo{}
    // 	if err := c.BodyParser(todo); err != nil {
    // 		return err
    // 	}
    // 	if todo.Body == "" {
    // 		return c.Status(400).JSON(fiber.Map{"error": "Body is required"})
    // 	}
    // 	todo.ID = len(todos) + 1
    // 	todos = append(todos, *todo)
    // 	return c.Status(201).JSON(todo)
    // })
    // // Update a todo
    // app.Patch("/api/todos/:id", func(c *fiber.Ctx) error {
    // 	id := c.Params("id")
    // 	for i, todo := range todos {
    // 		if fmt.Sprint(todo.ID) == id {
    // 			todos[i].Completed = true
    // 			return c.Status(200).JSON(todos[i])
    // 		}
    // 	}
    // 	return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
    // })
    // // Delete a todo
    // app.Delete("/api/todos/:id", func(c *fiber.Ctx) error {
    // 	id := c.Params("id")
    // 	for i, todo := range todos {
    // 		if fmt.Sprint(todo.ID) == id {
    // 			todos = append(todos[:i], todos[i+1:]...)
    // 			return c.Status(200).JSON(fiber.Map{"success": true})
    // 		}
    // 	}
    // 	return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
    // })
    // log.Fatal(app.Listen(":" + PORT))