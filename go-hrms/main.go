package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type MongoIntance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var mg MongoIntance

const dbName = "fiber-hrms"
const mongoURI = "mongodb://localhost:27017/"+ dbName

type Employee struct {
	ID     string  `json:"id,omitempty" bson:"_id,omitempty"`
	Name   string  `json:"name"`
	Salary float64 `json:"salary"`
	Age    float64 `json:"age"`
}

func Connect() error {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client.Connect(ctx)
	db := client.Database(dbName)

	if err != nil {
		return err
	}
	mg = MongoIntance{
		Client: client,
		Db:     db,
	}
	return nil
}

func cancel() {
	panic("unimplemented")
}

func main() {

	if err := Connect(); err != nil {
		log.Fatal(err)
	}
	app := fiber.New()

	app.Get("/employee", func(c *fiber.Ctx) error {
		query := bson.D{{}}
		cursor, err := mg.Db.Collection("employess").Find(c.Context(), query)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		var employees []Employee = make([]Employee, 0) //slice

		if err := cursor.All(c.Context(), &employees); err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.JSON(employees)

	})
	app.Post("/employee", func(c *fiber.Ctx) error {
		collection := mg.Db.Collection("employess")

		employee := new(Employee)
		if err := c.BodyParser(employee); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		r, err := collection.InsertOne(c.Context(), employee)

		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		filter := bson.D{{Key: "_id", Value: r.InsertedID}}

		createdRecord := collection.FindOne(c.Context(), filter)

		createdEmployee := &Employee{}

		createdRecord.Decode(createdEmployee)

		return c.Status(201).JSON(createdEmployee)

	})
	//update
	app.Put("/employee/:id", (func(c *fiber.Ctx) error {
		idParam := c.Params("id")

		employeeId, err := primitive.ObjectIDFromHex(idParam)

		if err != nil {
			return c.SendStatus(400)
		}

		employee := new(Employee)

		if err := c.BodyParser(employee); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		query := bson.D{{Key: "_id", Value: employeeId}}

		update := bson.D{
			{Key: "$set",
				Value: bson.D{
					{Key: "name", Value: employee.Name},
					{Key: "salary", Value: employee.Salary},
					{Key: "age", Value: employee.Age},
				},
			},
		}
		err = mg.Db.Collection("employess").FindOneAndUpdate(c.Context(), query, update).Err()

		if err != nil {
			if err ==  mongo.ErrNoDocuments{
				return c.SendStatus(400)
			}
			return c.SendStatus(500)
		}
		employee.ID = idParam
		return c.Status(200).JSON(employee)
	}))
	app.Delete("/employee/:id",func(c *fiber.Ctx) error {
		collection := mg.Db.Collection("employess")
		
		employeeId,err := primitive.ObjectIDFromHex(c.Params("id"))

		if err != nil {
			return c.SendStatus(400)
		}

		query := bson.D{{Key:"_id",Value:employeeId}}
		result,err := collection.DeleteOne(c.Context(),query)

		if err != nil {
			return c.SendStatus(500)
		}
		if result.DeletedCount <1 {
			return c.SendStatus(404)
		}
		return c.Status(200).JSON("record deleted")

	})

	log.Fatal(app.Listen(":3000"))
}
