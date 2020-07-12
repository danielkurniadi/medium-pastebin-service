package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/iqdf/pastebin-service/domain"
	pasteHTTPLib "github.com/iqdf/pastebin-service/paste/delivery/http"
	pasteMongoLib "github.com/iqdf/pastebin-service/paste/repository/mongo"
	pasteServiceLib "github.com/iqdf/pastebin-service/paste/service"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	var (
		dbConn *mongo.Client

		pasteRepo    domain.PasteRepository
		pasteService domain.PasteService

		rootRouter  *mux.Router
		pasteRouter *mux.Router
	)

	// Setup database connection here...
	ctx, cancelConn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelConn()

	dbOpt := options.Client().ApplyURI("mongodb://localhost:27017/" + "pastebinDB")
	dbConn, connErr := mongo.Connect(ctx, dbOpt)

	if connErr != nil {
		panic("[app/main] entrypoint unable to connect to database: " + connErr.Error())
	}

	// Setup repositories here ...
	pasteRepo = pasteMongoLib.NewPasteRepo(dbConn, "pastebinDB")

	// Setup services here ...
	pasteService = pasteServiceLib.NewPasteService(pasteRepo)

	// Setup middlewares here ...
	dummyMiddleware := func(h http.Handler) http.Handler { return h }

	// Register routers here ...
	rootRouter = mux.NewRouter()
	pasteRouter = rootRouter.PathPrefix("/").Subrouter()

	pasteHTTPLib.NewPasteHandler(pasteService).Routes(pasteRouter, dummyMiddleware)

	server := &http.Server{
		Addr:         "0.0.0.0:8080", // equivalent to localhost with port: 8080
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 15,
		Handler:      rootRouter,
	}

	// Starting server and listening
	fmt.Println("Starting server and listening at localhost:8000 ...")
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	// Handle shutdowns gracefully at SIGINT (Ctrl + C) interupts
	// Note: SIGKILL, SIGQUIT, or SIGTERM will not be caught
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// block until we receive SIGINT
	<-c

	ctx, cancelShutdown := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelShutdown()

	server.Shutdown(ctx)

	log.Println("Shutting down ....")
	os.Exit(0)
}
