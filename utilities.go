package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp_helper "github.com/CodeClarityCE/utility-amqp-helper"
	dbhelper "github.com/CodeClarityCE/utility-dbhelper/helper"
	types_amqp "github.com/CodeClarityCE/utility-types/amqp"
	types_analysis "github.com/CodeClarityCE/utility-types/analysis"
	types_plugin "github.com/CodeClarityCE/utility-types/plugin"
	"github.com/arangodb/go-driver"
)

// callback is a function that processes a message received from a plugin dispatcher.
// It takes the following parameters:
// - args: any, the arguments passed to the callback function.
// - config: types_plugin.Plugin, the configuration of the plugin.
// - message: []byte, the message received from the plugin dispatcher.
//
// The callback function performs the following steps:
// 1. Extracts the arguments from the args parameter.
// 2. Opens a database connection.
// 3. Reads the message and unmarshals it into a dispatcherMessage struct.
// 4. Starts a timer to measure the execution time.
// 5. Retrieves the analysis document from the database.
// 6. Starts the analysis using the startAnalysis function.
// 7. Prints the elapsed time.
// 8. Updates the analysis with the results and status.
// 9. Commits the transaction.
// 10. Sends the results to the plugins_dispatcher.
//
// If any error occurs during the execution of the callback function, it will be logged and the transaction will be aborted.
func callback(args any, config types_plugin.Plugin, message []byte) {
	// Get arguments
	s, ok := args.(Arguments)
	if !ok {
		log.Printf("not ok")
		return
	}

	// Read message
	var dispatcherMessage types_amqp.DispatcherPluginMessage
	err := json.Unmarshal([]byte(message), &dispatcherMessage)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	// Start timer
	start := time.Now()

	// Open DB
	db, tctx, trxid, err := dbhelper.OpenDatabase(dbhelper.Config.Database.Results)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	// Get analysis
	analysis_document, _, err := dbhelper.GetAnalysis(db, tctx, dispatcherMessage.AnalysisId)
	if err != nil {
		log.Printf("%v", err)
		db.AbortTransaction(tctx, trxid, nil)
		return
	}

	// Commit transaction
	err = db.CommitTransaction(tctx, trxid, nil)
	if err != nil {
		log.Printf("Failed to commit transaction: %t", err)
		return
	}

	// Start analysis
	result, status, err := startAnalysis(s, dispatcherMessage, config, analysis_document)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	// Print time elapsed
	t := time.Now()
	elapsed := t.Sub(start)
	log.Println(elapsed)

	// Open DB
	db, tctx, trxid, err = dbhelper.OpenDatabase(dbhelper.Config.Database.Results)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	// Get analysis
	analysis_document, analysis_col, err := dbhelper.GetAnalysis(db, tctx, dispatcherMessage.AnalysisId)
	if err != nil {
		log.Printf("%v", err)
		db.AbortTransaction(tctx, trxid, nil)
		return
	}

	// Send results
	err = updateAnalysis(tctx, analysis_col, result, status, analysis_document, dispatcherMessage.AnalysisId, config, start, t)
	if err != nil {
		log.Printf("%v", err)
		db.AbortTransaction(tctx, trxid, nil)
		return
	}

	// Commit transaction
	err = db.CommitTransaction(tctx, trxid, nil)
	if err != nil {
		log.Printf("Failed to commit transaction: %t", err)
		return
	}

	// Send results
	sbom_message := types_amqp.PluginDispatcherMessage{
		AnalysisId: dispatcherMessage.AnalysisId,
		Plugin:     config.Name,
	}
	data, _ := json.Marshal(sbom_message)
	amqp_helper.Send("plugins_dispatcher", data)
}

// readConfig reads the configuration file and returns a Plugin object and an error.
// The configuration file is expected to be named "config.json" and should be located in the same directory as the source file.
// If the file cannot be opened or if there is an error decoding the file, an error is returned.
// The returned Plugin object contains the parsed configuration values, with the Key field set as the concatenation of the Name and Version fields.
// If there is an error registering the plugin, an error is returned.
func readConfig() (types_plugin.Plugin, error) {
	// Read config file
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Printf("%v", err)
		return types_plugin.Plugin{}, err
	}
	defer configFile.Close()

	// Decode config file
	var config types_plugin.Plugin
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		log.Printf("%v", err)
		return types_plugin.Plugin{}, err
	}
	config.Key = config.Name + ":" + config.Version

	err = register(config)
	if err != nil {
		log.Printf("%v", err)
		return types_plugin.Plugin{}, err
	}

	return config, nil
}

// register is a function that registers a plugin configuration in the database.
// It takes a config parameter of type types_plugin.Plugin, which represents the plugin configuration to be registered.
// The function returns an error if there was an issue with the registration process.
func register(config types_plugin.Plugin) error {
	db, err := dbhelper.GetDatabase(dbhelper.Config.Database.Plugins)
	if err != nil {
		log.Printf("%v", err)
		db, err = dbhelper.CreateDatabase(dbhelper.Config.Database.Plugins)
		if err != nil {
			log.Printf("%v", err)
			return err
		}
	}
	graph, err := dbhelper.GetOrCreatePluginsGraph(db)
	if err != nil {
		log.Printf("%v", err)
		return err
	}

	exists, err := dbhelper.CheckDocumentExists(db, "PLUGINS", config.Key)
	if err != nil {
		log.Printf("%v", err)
		return err
	}
	if !exists {
		_, err = dbhelper.AddDocument(graph, config, "PLUGINS")
		if err != nil {
			log.Printf("%v", err)
			return err
		}
	}
	return nil
}

// updateAnalysis updates the analysis document in the database with the provided result and status.
// It searches for the step with the same name as the plugin's name in the specified stage of the analysis document.
// If the step is found, its status is updated to "success" or "failure" based on the provided status.
// The result is stored in the step's result field.
// Finally, the updated analysis document is saved back to the database.
// If the step is not found, an error is returned.
func updateAnalysis(tctx context.Context, analyses_col driver.Collection, result map[string]any, status types_analysis.AnalysisStatus, analysis_document types_analysis.Analysis, analysis_id string, config types_plugin.Plugin, start time.Time, end time.Time) error {
	for step_id, step := range analysis_document.Steps[analysis_document.Stage] {
		// Update step status
		if step.Name == config.Name {
			analysis_document.Steps[analysis_document.Stage][step_id].Status = "success"
			if status == types_analysis.FAILURE {
				analysis_document.Steps[analysis_document.Stage][step_id].Status = "failure"
			}

			analysis_document.Steps[analysis_document.Stage][step_id].Result = make(map[string]any)
			analysis_document.Steps[analysis_document.Stage][step_id].Result = result

			analysis_document.Steps[analysis_document.Stage][step_id].Started_on = start.Format(time.RFC3339Nano)
			analysis_document.Steps[analysis_document.Stage][step_id].Ended_on = end.Format(time.RFC3339Nano)

			_, err := analyses_col.UpdateDocument(tctx, analysis_id, analysis_document)
			if err != nil {
				log.Printf("%v", err)
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("step not found")
}
