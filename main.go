package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/FrMnJ/testbot/internal/config"
	"github.com/FrMnJ/testbot/pkg/testbot"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Println("Error loading .env file:", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Println("Failed to load config:", err)
		return
	}

	g := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithDefaultModel(cfg.Model),
	)

	messageGeneratorFlow := genkit.DefineFlow(g, "messageGeneratorFlow", func(ctx context.Context, input *testbot.GenerateMessage) (*testbot.Messsage, error) {
		prompt := fmt.Sprintf(`Generate a message given that you are a user in chatbot app with capabilities to read data from the app. 
		Such as read docs, read tramites and read processes. The message should be based on the following scenario:
Actor: %s
Action: The user wants to ask %s

The message should be a natural language message that a user would send in the chatbot app based on the given scenario. In the language of the actor and action.`, input.Roles[0], input.Action)

		message, _, err := genkit.GenerateData[testbot.Messsage](ctx, g,
			ai.WithPrompt(prompt),
		)

		if err != nil {
			log.Println("Error generating message:", err)
			return nil, err
		}

		return message, nil
	})

	evaluatorFlow := genkit.DefineFlow(g, "evaluatorFlow", func(ctx context.Context, scenario *testbot.ScenarioDefinition) (*testbot.Score, error) {
		prompt := fmt.Sprintf(`You are an evaluator for a chatbot application.

Evaluate the response produced by the assistant based on the given scenario.

Scenario:
Actor: %s
Action: The user wants to ask %s

User Message:
%s

Assistant Response:
%s

The model made tool calls: %v

Evaluation Instructions:

Score the response on a scale from 0 to 5:

0 = Completely incorrect or irrelevant  
1 = Mostly incorrect  
2 = Partially correct but with major issues  
3 = Neutral / information not available  
4 = Good response with minor issues  
5 = Perfect response

Important rules:

If the assistant indicates that the information was not found or says something equivalent to "No cuento con suficiente información para responder eso", assign a score of **3 (neutral)**.

If the model made tool calls but the response does not yet contain the final information, **do not penalize the score**. In this case assign a score of **3**.

Provide concise feedback explaining:
- Why the score was given
- How the response could be improved

Write the feedback in the same language used by the actor and action.`, scenario.Roles[0], scenario.Action, scenario.Message, scenario.Response, scenario.CallTool)

		score, _, err := genkit.GenerateData[testbot.Score](ctx, g,
			ai.WithPrompt(prompt),
		)

		if err != nil {
			log.Println("Error generating score:", err)
			return nil, err
		}

		return score, nil
	})

	reviewerFlow := testbot.NewReviewerFlow(ctx, g, messageGeneratorFlow, evaluatorFlow, testbot.SendMessage, cfg)

	scenarios, err := testbot.ReadScenarios(cfg)
	if err != nil {
		log.Println("Failed to read scenarios:", err)
		return
	}
	var results []testbot.ResultScenario

	for i := 0; i < len(scenarios); i++ {
		log.Printf("Scenario: %s \nActor: %s", scenarios[i].Action, scenarios[i].Roles[0])
		if err := ctx.Err(); err != nil {
			log.Println("Execution interrupted, stopping scenarios:", err)
			break
		}

		score, err := reviewerFlow.Run(ctx, &scenarios[i])
		if err != nil {
			if ctx.Err() != nil {
				log.Println("Execution interrupted while running scenario")
				break
			}
			log.Printf("Error running scenario '%s': %v\n", scenarios[i].Action, err)
			continue
		}
		results = append(results, testbot.ResultScenario{
			ID:        scenarios[i].ID,
			Question:  scenarios[i].Message,
			Response:  scenarios[i].Response,
			Score:     score.Score,
			IsSuccess: score.Score >= 3,
			Feedback:  score.Feedback,
			CallTool:  scenarios[i].CallTool,
		})
		log.Printf("Score for scenario '%s': %d\n", scenarios[i].Action, score.Score)
	}

	err = testbot.SaveResultScenarios(results)
	if err != nil {
		log.Println("Error saving scenarios:", err)
	}
}
