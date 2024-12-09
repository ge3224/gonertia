package gonertia

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

// Inertia is a main Gonertia structure, which contains all the logic for being an Inertia adapter.
type Inertia struct {
	rootTemplate *template.Template

	sharedProps        Props
	sharedTemplateData TemplateData

	flash FlashProvider

	ssrURL        string
	ssrHTTPClient *http.Client

	containerID    string
	version        string
	encryptHistory bool
	jsonMarshaller JSONMarshaller
	logger         Logger
}

// New creates an Inertia instance with a root HTML template.
//
// Parameters:
//   - rootTemplateHTML: The base HTML template containing Inertia placeholders.
//   - funcMap: Optional template function map for custom template functions.
//   - opts: Optional configuration options for Inertia initialization.
func New(rootTemplateHTML string, funcMap *TemplateFuncs, opts ...Option) (*Inertia, error) {
	if rootTemplateHTML == "" {
		return nil, fmt.Errorf("blank root template")
	}

	tmpl := template.New("root")

	if funcMap != nil {
		tmpl.Funcs(template.FuncMap(*funcMap))
	}

	tmpl, err := tmpl.Parse(rootTemplateHTML)
	if err != nil {
		return nil, fmt.Errorf("parse root template: %w", err)
	}

	return NewFromTemplate(tmpl, opts...)
}

// NewFromFile reads all bytes from the root template file and then initializes Inertia.
func NewFromFile(rootTemplatePath string, funcMap *TemplateFuncs, opts ...Option) (*Inertia, error) {
	bs, err := os.ReadFile(rootTemplatePath)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", rootTemplatePath, err)
	}

	return NewFromBytes(bs, funcMap, opts...)
}

// NewFromReader reads all bytes from the reader with root template html and then initializes Inertia.
func NewFromReader(rootTemplateReader io.Reader, funcMap *TemplateFuncs, opts ...Option) (*Inertia, error) {
	bs, err := io.ReadAll(rootTemplateReader)
	if err != nil {
		return nil, fmt.Errorf("read root template: %w", err)
	}
	if closer, ok := rootTemplateReader.(io.Closer); ok {
		_ = closer.Close()
	}

	return NewFromBytes(bs, funcMap, opts...)
}

// NewFromBytes receive bytes with root template html and then initializes Inertia.
func NewFromBytes(rootTemplateBs []byte, funcMap *TemplateFuncs, opts ...Option) (*Inertia, error) {
	return New(string(rootTemplateBs), funcMap, opts...)
}

// NewFromTemplate receives a *template.Template and then initializes Inertia.
func NewFromTemplate(rootTemplate *template.Template, opts ...Option) (*Inertia, error) {
	if rootTemplate == nil {
		return nil, fmt.Errorf("nil root template")
	}

	i := &Inertia{
		rootTemplate:       rootTemplate,
		jsonMarshaller:     jsonDefaultMarshaller{},
		containerID:        "app",
		logger:             log.New(io.Discard, "", 0),
		sharedProps:        make(Props),
		sharedTemplateData: make(TemplateData),
	}

	for _, opt := range opts {
		if err := opt(i); err != nil {
			return nil, fmt.Errorf("initialize inertia: %w", err)
		}
	}

	return i, nil
}

// Logger defines an interface for debug messages.
type Logger interface {
	Printf(format string, v ...any)
	Println(v ...any)
}

// FlashProvider defines an interface for a flash data provider.
type FlashProvider interface {
	FlashErrors(ctx context.Context, errors ValidationErrors) error
	GetErrors(ctx context.Context) (ValidationErrors, error)
	ShouldClearHistory(ctx context.Context) (bool, error)
	FlashClearHistory(ctx context.Context) error
}

// ShareProp adds passed prop to shared props.
func (i *Inertia) ShareProp(key string, val any) {
	i.sharedProps[key] = val
}

// SharedProps returns shared props.
func (i *Inertia) SharedProps() Props {
	return i.sharedProps
}

// SharedProp return the shared prop.
func (i *Inertia) SharedProp(key string) (any, bool) {
	val, ok := i.sharedProps[key]
	return val, ok
}

// ShareTemplateData adds passed data to shared template data.
func (i *Inertia) ShareTemplateData(key string, val any) {
	i.sharedTemplateData[key] = val
}
