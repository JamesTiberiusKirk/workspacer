package cli

import (
	"flag"
	"strconv"
)

// FlagValues holds parsed flag values and provides type-safe accessors
type FlagValues struct {
	flagSet *flag.FlagSet
}

// ParseFlags creates a FlagSet from flag declarations and parses the args
func ParseFlags(flagDefs []Flag, args []string) *FlagValues {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	// Register all flags
	for _, f := range flagDefs {
		switch f.Type {
		case "bool":
			defaultVal := f.Default == "true"
			fs.Bool(f.Name, defaultVal, f.Description)
			if f.Short != "" {
				fs.Bool(f.Short, defaultVal, f.Description)
			}
		case "string":
			fs.String(f.Name, f.Default, f.Description)
			if f.Short != "" {
				fs.String(f.Short, f.Default, f.Description)
			}
		case "int":
			defaultVal, _ := strconv.Atoi(f.Default)
			fs.Int(f.Name, defaultVal, f.Description)
			if f.Short != "" {
				fs.Int(f.Short, defaultVal, f.Description)
			}
		}
	}

	// Parse the args
	fs.Parse(args)

	return &FlagValues{flagSet: fs}
}

// Bool returns the boolean value of a flag
func (fv *FlagValues) Bool(name string) bool {
	if f := fv.flagSet.Lookup(name); f != nil {
		if val, ok := f.Value.(interface{ Get() interface{} }); ok {
			if b, ok := val.Get().(bool); ok {
				return b
			}
		}
	}
	return false
}

// String returns the string value of a flag
func (fv *FlagValues) String(name string) string {
	if f := fv.flagSet.Lookup(name); f != nil {
		return f.Value.String()
	}
	return ""
}

// Int returns the int value of a flag
func (fv *FlagValues) Int(name string) int {
	if f := fv.flagSet.Lookup(name); f != nil {
		if val, ok := f.Value.(interface{ Get() interface{} }); ok {
			if i, ok := val.Get().(int); ok {
				return i
			}
		}
	}
	return 0
}

// Args returns the remaining non-flag arguments
func (fv *FlagValues) Args() []string {
	return fv.flagSet.Args()
}

// GetBoolFlag returns a boolean flag value from the context
// Panics if the flag doesn't exist in the command's flag definitions
func GetBoolFlag(ctx ConfigMapCtx, name string) bool {
	if ctx.ParsedFlags == nil {
		panic("no flags parsed for this command")
	}
	return ctx.ParsedFlags.Bool(name)
}

// GetStringFlag returns a string flag value from the context
// Panics if the flag doesn't exist in the command's flag definitions
func GetStringFlag(ctx ConfigMapCtx, name string) string {
	if ctx.ParsedFlags == nil {
		panic("no flags parsed for this command")
	}
	return ctx.ParsedFlags.String(name)
}

// GetIntFlag returns an int flag value from the context
// Panics if the flag doesn't exist in the command's flag definitions
func GetIntFlag(ctx ConfigMapCtx, name string) int {
	if ctx.ParsedFlags == nil {
		panic("no flags parsed for this command")
	}
	return ctx.ParsedFlags.Int(name)
}

// GetFlagArgs returns the remaining non-flag arguments
func GetFlagArgs(ctx ConfigMapCtx) []string {
	if ctx.ParsedFlags == nil {
		return ctx.Args
	}
	return ctx.ParsedFlags.Args()
}
