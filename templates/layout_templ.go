// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.793
package templates

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"github.com/datasektionen/sso/pkg/config"
	"github.com/datasektionen/sso/pkg/static"
)

func base() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<!doctype html><html lang=\"en\"><head><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><title>Datasektionen SSO</title><link rel=\"stylesheet\" href=\"/public/style.dist.css\"><script src=\"https://unpkg.com/htmx.org@2.0.3\" integrity=\"sha384-0895/pl2MU10Hqc6jd4RvrthNlDiE9U1tWmX7WRESftEDRosgxNsQG/Ze9YMRzHq\" crossorigin=\"anonymous\"></script><script src=\"https://unpkg.com/htmx-ext-sse@2.2.2/sse.js\" integrity=\"sha384-fw+eTlCc7suMV/1w/7fr2/PmwElUIt5i82bi+qTiLXvjRXZ2/FkiTNA/w0MhXnGI\" crossorigin=\"anonymous\"></script><script src=\"https://unpkg.com/hyperscript.org@0.9.13\" integrity=\"sha384-5yQ5JTatiFEgeiEB4mfkRI3oTGtaNpbJGdcciZ4IEYFpLGt8yDsGAd7tKiMwnX9b\" crossorigin=\"anonymous\"></script><script src=\"https://unpkg.com/idiomorph@0.3.0/dist/idiomorph-ext.min.js\" integrity=\"sha384-01awMgY2Qxoo57dFZwehcB4wqi9TunC6fiF9hpPaDsLu+ayOG+WvoatvgPWquZh8\" crossorigin=\"anonymous\"></script><!--script src=\"https://unpkg.com/morphdom@2.7.4/dist/morphdom-esm.js\" integrity=\"sha384-32uSq+5CkJEgC7ZSxezdvE0+W6PeidDHEZi4Nh/Wm7EkCDcULeQVKes4llhG/Yyl\" crossorigin=\"anonymous\" type=\"module\"></script-->")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = base64Helpers().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if config.Config.Dev {
			templ_7745c5c3_Err = devAutoReload().Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</head><body class=\"bg-gray-900 text-neutral-100\" hx-ext=\"morph\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ_7745c5c3_Var1.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</body></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

func page(navItems templ.Component) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var2 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var2 == nil {
			templ_7745c5c3_Var2 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Var3 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
			templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
			templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
			if !templ_7745c5c3_IsBuffer {
				defer func() {
					templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
					if templ_7745c5c3_Err == nil {
						templ_7745c5c3_Err = templ_7745c5c3_BufErr
					}
				}()
			}
			ctx = templ.InitializeContext(ctx)
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<nav class=\"\n\t\t\t\tw-full h-12 bg-gray-700 flex uppercase\n\t\t\t\toverflow-auto\n\t\t\t\t[&amp;&gt;a:not(:first-child)]:flex [&amp;&gt;a]:items-center [&amp;&gt;a:not(:first-child)]:px-3 hover:[&amp;&gt;a]:bg-gray-600 [&amp;&gt;a]:transition-colors\n\t\t\t\t[&amp;&gt;a]:shrink-0\n\t\t\t\"><a href=\"/\" class=\"h-full block aspect-square grid place-items-center fill-cerise-regular bg-gray-600\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templ.Raw(static.PublicAsString("delta.svg")).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</a>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = navItems.Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</nav><div class=\"flex justify-center min-h-[calc(100vh-3rem)]\"><main class=\"bg-gray-800 w-[1000px] md:max-w-[calc(100vw-8rem)]\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templ_7745c5c3_Var2.Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</main></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return templ_7745c5c3_Err
		})
		templ_7745c5c3_Err = base().Render(templ.WithChildren(ctx, templ_7745c5c3_Var3), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

func nav(includeAdmin bool) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var4 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var4 == nil {
			templ_7745c5c3_Var4 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<a href=\"/account\">Account</a> ")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if includeAdmin {
			templ_7745c5c3_Err = adminNav().Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<a href=\"/logout\">Log out</a>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

func modal() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var5 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var5 == nil {
			templ_7745c5c3_Var5 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Var6 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
			templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
			templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
			if !templ_7745c5c3_IsBuffer {
				defer func() {
					templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
					if templ_7745c5c3_Err == nil {
						templ_7745c5c3_Err = templ_7745c5c3_BufErr
					}
				}()
			}
			ctx = templ.InitializeContext(ctx)
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"grid place-items-center min-h-screen\"><main class=\"bg-gray-800 rounded w-[22rem] max-w-[calc(100vw-2rem)]\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templ_7745c5c3_Var5.Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</main></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return templ_7745c5c3_Err
		})
		templ_7745c5c3_Err = base().Render(templ.WithChildren(ctx, templ_7745c5c3_Var6), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var input = `
	border border-neutral-500 grow
	outline-none focus:border-cerise-strong hover:border-cerise-light
	bg-slate-800 p-1.5 rounded leading-tight
`

var button = `
	bg-[#3f4c66] p-1 leading-tight block rounded border text-center
	select-none border-transparent outline-none
	enabled:focus:border-cerise-strong enabled:hover:border-cerise-light
	disabled:text-gray-500
`

var roundButton = `
	bg-[#3f4c66] shrink-0 h-5 w-5 rounded-full
	grid place-items-center pointer
	border border-transparent outline-none focus:border-cerise-strong hover:border-cerise-light relative
	[&>img]:w-3/5 [&>img]:h-3/5 [&>img]:invert
	[&>p]:text-center [&>p]:leading-none
`

var selectStyle = `
	border border-neutral-500
	outline-none focus:border-cerise-strong hover:border-cerise-light
	bg-slate-800 p-1.5 rounded leading-tight
`

func base64Helpers() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var7 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var7 == nil {
			templ_7745c5c3_Var7 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<script>\n/**\n * @param {string} data\n * @returns Uint8Array\n */\nlet decodebase64url = (data) =>\n    Uint8Array.from(window.atob(data.replace(/-/g, \"+\").replace(/_/g, \"/\")), v => v.charCodeAt(0));\n/**\n * @param {Uint8Array} data\n * @returns string\n */\nlet encodebase64url = (data) =>\n    base64ArrayBuffer(data).replace(/\\+/g, \"-\").replace(/\\//g, \"_\");\n\n// Taken from https://gist.github.com/jonleighton/958841 and slightly modified.\n/*\nMIT LICENSE\nCopyright 2011 Jon Leighton\nPermission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the \"Software\"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:\nThe above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.\nTHE SOFTWARE IS PROVIDED \"AS IS\", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.\n*/\n\n/**\n * @param {ArrayBuffer} arrayBuffer\n */\nfunction base64ArrayBuffer(arrayBuffer) {\n    let base64 = \"\";\n    let encodings = \"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/\";\n\n    let bytes = new Uint8Array(arrayBuffer);\n    let byteLength = bytes.byteLength;\n    let byteRemainder = byteLength % 3;\n    let mainLength = byteLength - byteRemainder;\n\n    // Main loop deals with bytes in chunks of 3\n    for (let i = 0; i < mainLength; i = i + 3) {\n        // Combine the three bytes into a single integer\n        let chunk = (bytes[i] << 16) | (bytes[i + 1] << 8) | bytes[i + 2];\n\n        // Use bitmasks to extract 6-bit segments from the triplet\n        let a = (chunk & 16515072) >> 18; // 16515072 = (2^6 - 1) << 18\n        let b = (chunk & 258048) >> 12; // 258048   = (2^6 - 1) << 12\n        let c = (chunk & 4032) >> 6; // 4032     = (2^6 - 1) << 6\n        let d = chunk & 63;               // 63       = 2^6 - 1\n\n        // Convert the raw binary segments to the appropriate ASCII encoding\n        base64 += encodings[a] + encodings[b] + encodings[c] + encodings[d];\n    }\n\n    // Deal with the remaining bytes and padding\n    if (byteRemainder == 1) {\n        let chunk = bytes[mainLength];\n\n        let a = (chunk & 252) >> 2; // 252 = (2^6 - 1) << 2\n\n        // Set the 4 least significant bits to zero\n        let b = (chunk & 3) << 4; // 3   = 2^2 - 1\n\n        base64 += encodings[a] + encodings[b] + '==';\n    } else if (byteRemainder == 2) {\n        let chunk = (bytes[mainLength] << 8) | bytes[mainLength + 1];\n\n        let a = (chunk & 64512) >> 10; // 64512 = (2^6 - 1) << 10\n        let b = (chunk & 1008) >> 4; // 1008  = (2^6 - 1) << 4\n\n        // Set the 2 least significant bits to zero\n        let c = (chunk & 15) << 2; // 15    = 2^4 - 1\n\n        base64 += encodings[a] + encodings[b] + encodings[c] + '=';\n    }\n\n    return base64;\n}\n\t</script>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate
