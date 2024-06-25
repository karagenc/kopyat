package kopyaship

import "github.com/tomruk/kopyaship/internal/scripting/ctx"

type (
	Context                = ctx.Context
	BackupContext          = ctx.BackupContext
	IfileGenerationContext = ctx.IfileGenerationContext
)

func GetContext() Context { panic("not implemented here") }
