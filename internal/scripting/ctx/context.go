package ctx

type Context interface {
	Backup() (c *BackupContext, ok bool)
	IfileGeneration() (c *IfileGenerationContext, ok bool)
}

type (
	BackupContext struct {
		// Is it a "pre" hook?
		Pre bool
		// Is it a "post" hook?
		Post bool

		Name string
		// Target path of backup. (e.g. path of the restic repository.)
		TargetPath string
		Base       string
		Paths      []string
		// Only valid for pre hook. If called within a post hook, this is no-op.
		Skip func()
		// Are we going to generate an ifile and use it?
		UseIfile bool
	}

	IfileGenerationContext struct {
		// Is it a "pre" hook?
		Pre bool
		// Is it a "post" hook?
		Post  bool
		Ifile string
		Type  string
	}
)

type context struct {
	backupContext          *BackupContext
	ifileGenerationContext *IfileGenerationContext
}

func NewEmptyContext() Context { return &context{} }

func NewBackupContext(
	pre bool,
	name string,
	targetPath string,
	base string,
	paths []string,
	skip func(),
	useIfile bool,
) Context {
	return &context{
		backupContext: &BackupContext{
			Pre:        pre,
			Post:       !pre,
			Name:       name,
			TargetPath: targetPath,
			Base:       base,
			Paths:      paths,
			Skip:       skip,
			UseIfile:   useIfile,
		},
	}
}

func NewIfileGenerationContext(pre bool, ifile string, typ string) Context {
	return &context{
		ifileGenerationContext: &IfileGenerationContext{
			Pre:   pre,
			Post:  !pre,
			Ifile: ifile,
			Type:  typ,
		},
	}
}

func (_c *context) Backup() (c *BackupContext, ok bool) {
	if _c.backupContext != nil {
		return _c.backupContext, true
	}
	return nil, false
}

func (_c *context) IfileGeneration() (c *IfileGenerationContext, ok bool) {
	if _c.ifileGenerationContext != nil {
		return _c.ifileGenerationContext, true
	}
	return nil, false
}
