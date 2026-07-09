package todorpc

import (
	"strconv"

	todov1 "github.com/minty1202/go-ddd-onion-template/gen/todo/v1"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/todo/define"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/todo/view"
)

func toDefineParam(req *todov1.DefineRequest) define.Param {
	return define.Param{
		Title: req.GetTitle(),
		Body:  req.GetBody(),
	}
}

func toViewParam(req *todov1.ViewRequest) view.Param {
	return view.Param{
		ID: req.GetId(),
	}
}

func toDefineResponse(r *define.Result) *todov1.DefineResponse {
	return &todov1.DefineResponse{
		Todo: toTodoProto(r.ID, r.Title, r.Body, r.Completed, r.Version),
	}
}

func toViewResponse(r *view.Result) *todov1.ViewResponse {
	return &todov1.ViewResponse{
		Todo: toTodoProto(r.ID, r.Title, r.Body, r.Completed, r.Version),
	}
}

func toTodoProto(id, title, body string, completed bool, version int) *todov1.Todo {
	return &todov1.Todo{
		Id:        id,
		Title:     title,
		Body:      body,
		Completed: completed,
		Etag:      strconv.Itoa(version),
	}
}
