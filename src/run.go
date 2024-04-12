package codeclarity

import (
	"log"

	"github.com/arangodb/go-driver"
)

// Entrypoint for the plugin
func Start(database driver.Database) any {
	// Start the plugin
	log.Println("Starting plugin...")
	log.Println("You use the database: ", database.Name())
	return "Hello, World!"
}
