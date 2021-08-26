package executor

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/kelseyhightower/envconfig"
	"github.com/kubeshop/kubtest-executor-template/internal/pkg/repository/result"
	"github.com/kubeshop/kubtest-executor-template/internal/pkg/worker"

	// TODO move server to kubtest/pkg
	"github.com/kubeshop/kubtest-executor-template/internal/pkg/server"

	"github.com/kubeshop/kubtest/pkg/api/kubtest"
)

// ConcurrentExecutions per node
const ConcurrentExecutions = 4

// NewExecutor returns new TemplateExecutor instance
func NewExecutor(resultRepository result.Repository) Executor {
	var httpConfig server.Config
	envconfig.Process("EXECUTOR", &httpConfig)

	e := Executor{
		HTTPServer: server.NewServer(httpConfig),
		Repository: resultRepository,
		Worker:     worker.NewWorker(resultRepository),
	}

	return e
}

type Executor struct {
	server.HTTPServer
	Repository result.Repository
	Worker     worker.Worker
}

func (p *Executor) Init() {
	executions := p.Routes.Group("/executions")
	executions.Post("/", p.StartExecution())
	executions.Get("/:id", p.GetExecution())
}

func (p *Executor) StartExecution() fiber.Handler {
	return func(c *fiber.Ctx) error {

		var request kubtest.ExecutionRequest
		err := json.Unmarshal(c.Body(), &request)
		if err != nil {
			return p.Error(c, http.StatusBadRequest, err)
		}

		execution := kubtest.NewExecution(string(request.Metadata), request.Params)
		err = p.Repository.Insert(context.Background(), execution)
		if err != nil {
			return p.Error(c, http.StatusInternalServerError, err)

		}

		p.Log.Infow("starting new execution", "execution", execution)
		c.Response().Header.SetStatusCode(201)
		return c.JSON(execution)
	}
}

func (p Executor) GetExecution() fiber.Handler {
	return func(c *fiber.Ctx) error {
		execution, err := p.Repository.Get(context.Background(), c.Params("id"))
		if err != nil {
			return p.Error(c, http.StatusInternalServerError, err)
		}

		return c.JSON(execution)
	}
}

func (p Executor) Run() error {
	executionsQueue := p.Worker.PullExecutions()
	p.Worker.Run(executionsQueue)

	return p.HTTPServer.Run()
}
