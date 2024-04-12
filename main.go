package main

import (
	"log"

	plugin "github.com/CodeClarityLU/plugin-plugin-name/src"

	amqp_helper "github.com/CodeClarityLU/amqp-helper"
	dbhelper "github.com/CodeClarityLU/dbhelper/helper"
	types_amqp "github.com/CodeClarityLU/types/amqp"
	types_analysis "github.com/CodeClarityLU/types/analysis"
	types_plugin "github.com/CodeClarityLU/types/plugin"
	"github.com/arangodb/go-driver"
)

// TODO: Define the arguments you want to pass to the callback function
// Usually a plugin needs access to at least one database
type Arguments struct {
	database  driver.Database
	knowledge driver.Database
	graph     driver.Graph
}

// main is the entry point of the program.
// It reads the configuration, initializes the necessary databases and graph,
// and starts listening on the queue.
func main() {
	config, err := readConfig()
	if err != nil {
		log.Printf("%v", err)
		return
	}

	// Define the arguments you want to pass to the callback function
	database, err := dbhelper.GetDatabase(dbhelper.Config.Database.Results)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	knowledge, err := dbhelper.GetDatabase(dbhelper.Config.Database.Knowledge)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	graph, err := dbhelper.GetGraph(dbhelper.Config.Database.Results, dbhelper.Config.Graph.Results)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	args := Arguments{
		database:  database,
		knowledge: knowledge,
		graph:     graph,
	}

	// Start listening on the queue
	amqp_helper.Listen("dispatcher_"+config.Name, callback, args, config)
}

// startAnalysis is a function that performs the analysis for codeclarity plugin.
// It takes the following parameters:
// - args: Arguments for the analysis.
// - dispatcherMessage: Dispatcher plugin message.
// - config: Plugin configuration.
// - analysis_document: Analysis document.
// It returns a map[string]any containing the result of the analysis, the analysis status, and an error if any.
func startAnalysis(args Arguments, dispatcherMessage types_amqp.DispatcherPluginMessage, config types_plugin.Plugin, analysis_document types_analysis.Analysis) (map[string]any, types_analysis.AnalysisStatus, error) {
	// Get analysis config
	// Reads the configuration of the plugin
	messageData := analysis_document.Config[config.Name].(map[string]any)

	// Prepare the arguments for the plugin
	aConfigAttribute := messageData["aConfigAttribute"].(string)
	log.Println(aConfigAttribute)

	// Get previous stage
	analysis_stage := analysis_document.Stage - 1
	log.Println(analysis_stage)

	// Do your analysis here
	output := plugin.Start(args.knowledge)

	// Add the result to the database
	documentKey, err := dbhelper.AddDocument(args.graph, output, "PLUGIN_NAME")
	if err != nil {
		log.Printf("%v", err)
		return nil, types_analysis.FAILURE, err
	}

	err = dbhelper.LinkDocuments(args.graph, dbhelper.Config.Collection.Analyses+"/"+dispatcherMessage.AnalysisId, dbhelper.Config.Collection.Sboms+"/"+documentKey, dbhelper.Config.Edge.Results)
	if err != nil {
		log.Printf("%v", err)
		return nil, types_analysis.FAILURE, err
	}

	// Prepare the result to store in step
	// Usually we store the key of the result document that was just created
	// The other plugins will use this key
	result := make(map[string]any)
	result["output"] = output

	// Set the status of the analysis
	status := types_analysis.SUCCESS

	// The output is always a map[string]any
	return result, status, nil
}
