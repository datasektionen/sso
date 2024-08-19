// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.747
package passkey

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"github.com/datasektionen/logout/pkg/templates"
	"github.com/datasektionen/logout/services/passkey/export"
	"github.com/go-webauthn/webauthn/protocol"
)

func passkeyLogin(kthid string, credAss *protocol.CredentialAssertion) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
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
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<form id=\"passkey-login-form\" hx-post=\"/login/passkey/begin\" hx-on::before-request=\"this.querySelectorAll(&#39;.error&#39;).forEach(e =&gt; e.remove())\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if credAss != nil {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" data-cred-ass=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var2 string
			templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(templ.JSONString(credAss))
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `passkey.templ`, Line: 15, Col: 44}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" hx-swap=\"outerHTML\" class=\"[&amp;&gt;.error]:bg-red-600/50 [&amp;&gt;.error]:p-2 [&amp;&gt;.error]:mt-2 [&amp;&gt;.error]:rounded\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if credAss != nil {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<script type=\"module\">\n\t\t\t\tlet form = document.querySelector(\"#passkey-login-form\");\n\t\t\t\tlet credAss = JSON.parse(form.dataset.credAss);\n\t\t\t\tcredAss.publicKey.challenge = decodebase64url(credAss.publicKey.challenge);\n\t\t\t\tfor (let ac of credAss.publicKey.allowCredentials) {\n\t\t\t\t\tac.id = decodebase64url(ac.id);\n\t\t\t\t}\n\t\t\t\tconsole.log(credAss);\n\t\t\t\tevent.preventDefault();\n\t\t\t\ttry {\n\t\t\t\t\tlet cred = await navigator.credentials.get(credAss);\n\t\t\t\t\tlet res = await fetch(\"/login/passkey/finish\", {\n\t\t\t\t\t\tmethod: \"post\",\n\t\t\t\t\t\theaders: { \"Content-Type\": \"application/json\" },\n\t\t\t\t\t\tbody: JSON.stringify({\n\t\t\t\t\t\t\tkthid: new FormData(form).get(\"kthid\"),\n\t\t\t\t\t\t\tcred: {\n\t\t\t\t\t\t\t\tid: cred.id,\n\t\t\t\t\t\t\t\trawId: encodebase64url(cred.rawId),\n\t\t\t\t\t\t\t\ttype: cred.type,\n\t\t\t\t\t\t\t\tauthenticatorAttachment: cred.authenticatorAttachment,\n\t\t\t\t\t\t\t\tresponse: {\n\t\t\t\t\t\t\t\t\tauthenticatorData: encodebase64url(cred.response.authenticatorData),\n\t\t\t\t\t\t\t\t\tclientDataJSON: encodebase64url(cred.response.clientDataJSON),\n\t\t\t\t\t\t\t\t\tsignature: encodebase64url(cred.response.signature),\n\t\t\t\t\t\t\t\t\tuserHandle: encodebase64url(cred.response.userHandle),\n\t\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t}),\n\t\t\t\t\t});\n\t\t\t\t\tif (res.status == 200)\n\t\t\t\t\t\twindow.location.replace(\"/\");\n\t\t\t\t\telse\n\t\t\t\t\t\tthrow new Error(await res.text());\n\t\t\t\t} catch (err) {\n\t\t\t\t\tlet text = (err.name === \"NotAllowedError\")\n\t\t\t\t\t\t? \"Missing permission or request was cancelled\"\n\t\t\t\t\t\t: err.message;\n\t\t\t\t\tlet el = document.createElement(\"p\");\n\t\t\t\t\tel.classList.add(\"error\");\n\t\t\t\t\tel.textContent = text;\n\t\t\t\t\tform.appendChild(el);\n\t\t\t\t} finally {\n\t\t\t\t\tform.querySelector(\"button\").classList.remove(\"spinner\");\n\t\t\t\t}\n\t\t\t</script>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<label class=\"text-sm\" for=\"pk-kthid\">Log In using a Passkey</label><div class=\"flex gap-2\"><input id=\"pk-kthid\" name=\"kthid\" type=\"text\" required placeholder=\"KTH ID\" value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var3 string
		templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(kthid)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `passkey.templ`, Line: 76, Col: 17}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" class=\"\n\t\t\t\t\tborder border-neutral-500 grow\n\t\t\t\t\toutline-none focus:border-cerise-strong hover:border-cerise-light\n\t\t\t\t\tbg-slate-800 p-1.5 rounded h-8\n\t\t\t\t\"> ")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var4 = []any{`
					bg-[#3f4c66] shrink-0 h-8 w-8 rounded-full
					grid place-items-center pointer
					border border-transparent outline-none focus:border-cerise-strong hover:border-cerise-light relative
				` + templates.BigIfTrue(credAss != nil, "spinner", "")}
		templ_7745c5c3_Err = templ.RenderCSSItems(ctx, templ_7745c5c3_Buffer, templ_7745c5c3_Var4...)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<button class=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var5 string
		templ_7745c5c3_Var5, templ_7745c5c3_Err = templ.JoinStringErrs(templ.CSSClasses(templ_7745c5c3_Var4).String())
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `passkey.templ`, Line: 1, Col: 0}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var5))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><img src=\"/public/key_icon.svg\" class=\"w-3/5 h-3/5 invert\"></button></div></form>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

func showPasskey(passkey export.Passkey) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
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
		templ_7745c5c3_Var6 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var6 == nil {
			templ_7745c5c3_Var6 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<li class=\"flex p-2 gap-2 items-center\"><span>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var7 string
		templ_7745c5c3_Var7, templ_7745c5c3_Err = templ.JoinStringErrs(passkey.Name)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `passkey.templ`, Line: 98, Col: 22}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var7))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</span> <button class=\"\n\t\t\t\tbg-[#3f4c66] shrink-0 h-5 w-5 rounded-full\n\t\t\t\tgrid place-items-center pointer\n\t\t\t\tborder border-transparent outline-none focus:border-cerise-strong hover:border-cerise-light relative\n\t\t\t\" hx-delete=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var8 string
		templ_7745c5c3_Var8, templ_7745c5c3_Err = templ.JoinStringErrs("/passkey/" + passkey.ID.String())
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `passkey.templ`, Line: 105, Col: 48}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var8))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" hx-target=\"closest li\" hx-swap=\"outerHTML\"><img class=\"w-3/5 h-3/5 invert\" src=\"/public/x.svg\"></button></li>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

func passkeySettings(passkeys []export.Passkey) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
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
		templ_7745c5c3_Var9 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var9 == nil {
			templ_7745c5c3_Var9 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<section class=\"flex flex-col\"><h2 class=\"text-lg\">Passkeys:</h2><ul id=\"passkey-list\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		for _, passkey := range passkeys {
			templ_7745c5c3_Err = showPasskey(passkey).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</ul><button hx-get=\"/passkey/add-form\" hx-swap=\"afterend\" hx-on:htmx:after-swap=\"this.style.display = &#39;none&#39;\" id=\"add-passkey-button\" class=\"\n\t\t\t\tbg-[#3f4c66] p-1.5 block rounded border text-center\n\t\t\t\tselect-none border-transparent outline-none\n\t\t\t\tfocus:border-cerise-strong hover:border-cerise-light\n\t\t\t\">Add passkey</button></section>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

func addPasskeyForm(cc *protocol.CredentialCreation) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
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
		templ_7745c5c3_Var10 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var10 == nil {
			templ_7745c5c3_Var10 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<form data-credential-creation=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var11 string
		templ_7745c5c3_Var11, templ_7745c5c3_Err = templ.JoinStringErrs(templ.JSONString(cc))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `passkey.templ`, Line: 138, Col: 49}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var11))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" onsubmit=\"addPasskey(this, event)\" class=\"[&amp;&gt;.error]:bg-red-600/50 [&amp;&gt;.error]:p-2 [&amp;&gt;.error]:mt-2 [&amp;&gt;.error]:rounded\"><script>\n\t\t\tasync function addPasskey(form, event) {\n\t\t\t\tevent.preventDefault();\n\t\t\t\tlet cc = JSON.parse(form.dataset.credentialCreation);\n\t\t\t\tcc.publicKey.challenge = decodebase64url(cc.publicKey.challenge);\n\t\t\t\tcc.publicKey.user.id = decodebase64url(cc.publicKey.user.id);\n\t\t\t\tfor (let err of form.querySelectorAll(\".error\"))\n\t\t\t\t\terr.remove();\n\n\t\t\t\ttry {\n\t\t\t\t\tlet cred = await navigator.credentials.create(await cc);\n\t\t\t\t\tlet res = await fetch(\"/passkey\", {\n\t\t\t\t\t\tmethod: \"post\",\n\t\t\t\t\t\theaders: { \"Content-Type\": \"application/json\" },\n\t\t\t\t\t\tbody: JSON.stringify({\n\t\t\t\t\t\t\tname: new FormData(form).get(\"name\"),\n\t\t\t\t\t\t\tid: cred.id,\n\t\t\t\t\t\t\ttype: cred.type,\n\t\t\t\t\t\t\tauthenticatorAttachment: cred.authenticatorAttachment,\n\t\t\t\t\t\t\tresponse: {\n\t\t\t\t\t\t\t\tattestationObject: encodebase64url(cred.response.attestationObject),\n\t\t\t\t\t\t\t\tclientDataJSON: encodebase64url(cred.response.clientDataJSON),\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t}),\n\t\t\t\t\t});\n\t\t\t\t\tif (res.status != 200)\n\t\t\t\t\t\tthrow new Error(await res.text());\n\t\t\t\t\tlet key = await res.text();\n\t\t\t\t\tform.remove();\n\t\t\t\t\thtmx.swap(\"#passkey-list\", key, { swapStyle: \"beforeend\" });\n\t\t\t\t\tdocument.querySelector(\"#add-passkey-button\").style.display = \"\";\n\t\t\t\t} catch (err) {\n\t\t\t\t\tlet text = (err.name === \"NotAllowedError\")\n\t\t\t\t\t\t? \"Missing permission or request was cancelled\"\n\t\t\t\t\t\t: err.message;\n\t\t\t\t\tlet el = document.createElement(\"p\");\n\t\t\t\t\tel.classList.add(\"error\");\n\t\t\t\t\tel.textContent = text;\n\t\t\t\t\tform.appendChild(el);\n\t\t\t\t}\n\t\t\t}\n\t\t</script><div class=\"flex gap-2\"><input placeholder=\"passkey name\" type=\"text\" autofocus name=\"name\" class=\"\n\t\t\t\t\tborder border-neutral-500 grow\n\t\t\t\t\toutline-none focus:border-cerise-strong hover:border-cerise-light\n\t\t\t\t\tbg-slate-800 p-1.5 rounded h-8\n\t\t\t\t\"> <button class=\"\n\t\t\t\tbg-[#3f4c66] shrink-0 h-8 w-8 rounded-full\n\t\t\t\tgrid place-items-center pointer\n\t\t\t\tborder border-transparent outline-none focus:border-cerise-strong hover:border-cerise-light\n\t\t\t\"><img class=\"w-3/5 h-3/5 invert\" src=\"/public/check.svg\"></button></div></form>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}
