import { For, createSignal } from "solid-js";
import { render } from "solid-js/web";
import { encodebase64url, decodebase64url } from "./base64";

function PasskeyLogin() {
    let [loading, setLoading] = createSignal(false);
    let [errors, setErrors] = createSignal<string[]>([]);
    let [kthid, setKthid] = createSignal("");

    async function submit(event: Event) {
        event.preventDefault();
        setLoading(true);
        setErrors([]);
        try {
            let beginRes = await fetch("/login/passkey/begin", {
                method: "post",
                body: kthid(),
            });
            if (beginRes.status != 200) throw new Error(await beginRes.text());
            let ca = await beginRes.json();
            ca.publicKey.challenge = decodebase64url(ca.publicKey.challenge);
            for (let c of ca.publicKey.allowCredentials) {
                c.id = decodebase64url(c.id);
            }
            // The ts definition did not seem to know what fields exist on a `Credential`
            let cred: any;
            try {
                cred = await navigator.credentials.get(ca);
            } catch (err) {
                throw new Error("Request denied or canceled");
            }
            let finishRes = await fetch("/login/passkey/finish", {
                method: "post",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    kthid: kthid(),
                    id: cred.id,
                    rawId: encodebase64url(cred.rawId),
                    type: cred.type,
                    authenticatorAttachment: cred.authenticatorAttachment,
                    response: {
                        authenticatorData: encodebase64url(cred.response.authenticatorData),
                        clientDataJSON: encodebase64url(cred.response.clientDataJSON),
                        signature: encodebase64url(cred.response.signature),
                        userHandle: encodebase64url(cred.response.userHandle),
                    },
                }),
            });
            if (finishRes.status == 200) window.location.replace("/");
        } catch (err) {
            if (err &&
                typeof err == "object" &&
                "message" in err &&
                typeof err.message == "string"
            )
                setErrors([...errors(), err.message]);
        } finally {
            setLoading(false);
        }
    }

    if (!("credentials" in navigator)) return;

    return <form onSubmit={submit} class="w-full">
        <label class="small" for="pk-kthid">Log In using a Passkey</label>
        <div class="row">
            <input
                id="pk-kthid"
                type="text"
                required
                placeholder="KTH ID"
                onInput={e => setKthid(e.target.value)}
            />
            <button
                disabled={loading()}
                class={"round-button " + (loading() ? "spinner" : "")}
            >
                <img src="/public/key_icon.svg" />
            </button>
        </div>
        <For each={errors()}>{(err) =>
            <p class="error">{err}</p>
        }</For>
    </form>;
}

render(PasskeyLogin, document.querySelector("#passkey-login")!);
