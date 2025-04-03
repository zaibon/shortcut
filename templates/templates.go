package templates

// import (
// 	"context"
// 	"embed"
// 	"strings"

// 	"github.com/donseba/go-htmx"

// 	"github.com/zaibon/shortcut/middleware"
// )

// //go:embed *.gohtml components/*.gohtml
// var fs embed.FS

// func Get(ctx context.Context, name string) htmx.RenderableComponent {
// 	if !strings.HasSuffix(name, ".gohtml") {
// 		name = name + ".gohtml"
// 	}
// 	page := htmx.NewComponent(name).
// 		FS(fs).
// 		Wrap(layout(ctx), "Content")

// 	for name, c := range components() {
// 		page.With(c, name)
// 	}

// 	return page
// }

// func layout(ctx context.Context) htmx.RenderableComponent {
// 	user := middleware.UserFromContext(ctx)
// 	isAutehnticated := user != nil

// 	navbarData := publicNavBar
// 	if isAutehnticated {
// 		navbarData = signedInNavBar
// 	}

// 	partials := map[string]htmx.RenderableComponent{
// 		"Navbar": htmx.NewComponent("navbar.gohtml").FS(fs).SetData(map[string]any{
// 			"Navbar": navbarData,
// 		}),
// 		"Footer": htmx.NewComponent("footer.gohtml").FS(fs),
// 	}

// 	layout := htmx.NewComponent("layout.gohtml").FS(fs)
// 	for i := range partials {
// 		layout.With(partials[i], i)
// 	}

// 	layout.SetGlobalData(map[string]any{
// 		"User": user,
// 	})

// 	return layout
// }

// func components() map[string]htmx.RenderableComponent {
// 	entries, err := fs.ReadDir("components")
// 	if err != nil {
// 		panic(err)
// 	}

// 	m := make(map[string]htmx.RenderableComponent, len(entries))

// 	for _, entry := range entries {
// 		if entry.IsDir() {
// 			continue
// 		}
// 		name := entry.Name()
// 		if strings.HasSuffix(name, ".gohtml") {
// 			name = strings.TrimSuffix(name, ".gohtml")
// 			key := strings.ToTitle(name[0:1]) + name[1:]
// 			m[key] = htmx.NewComponent("components/" + name + ".gohtml").FS(fs)
// 		}
// 	}

// 	return m
// }

// type navbarData struct {
// 	Name string
// 	Href string
// }

// var signedInNavBar = []navbarData{
// 	{Name: "Home", Href: "/"},
// 	{Name: "My URLs", Href: "/urls"},
// 	{Name: "Subscription", Href: "/subscription"},
// }

// var publicNavBar = []navbarData{
// 	{Name: "Home", Href: "/"},
// 	// {Name: "My URLs", Href: "/urls"},
// 	// {Name: "Subscription", Href: "/subscription"},
// }
