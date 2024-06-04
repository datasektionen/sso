import { For, Show, createSignal, onMount } from "solid-js";
import { render } from "solid-js/web";
import { z } from "zod";
import { decodebase64url, encodebase64url } from "./base64";

let passkeySchema = z.object({
    id: z.string(),
    name: z.string(),
});

type Passkey = z.infer<typeof passkeySchema>;

function PasskeySettings() {
    let [passkeys, setPasskeys] = createSignal<Passkey[]>([]);
    let [error, setError] = createSignal<string>("");
    let [loading, setLoading] = createSignal(true);
    let [adding, setAdding] = createSignal(false);

    onMount(async () => {
        let res = await fetch("/passkey/list");
        if (res.status != 200) {
            setError(await res.text());
            return;
        }
        let body = await res.json();
        setPasskeys(z.array(passkeySchema).parse(body));
        setLoading(false);
    });

    async function remove(id: string, index: number) {
        let res = await fetch("/passkey/remove", {
            method: "post",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(id),
        });
        if (res.status != 200) return setError(await res.text());
        let p = passkeys();
        setPasskeys([...p.slice(0, index), ...p.slice(index + 1)]);
    }

    return <section class="flex flex-col">
        <h2 class="text-lg">Passkeys:</h2>
        <ul>
            <Show when={loading()}>...</Show>
            <For each={passkeys()}>{(passkey, i) =>
                <li class="flex p-2 gap-2 items-center">
                    <span>{passkey.name}</span>
                    <button
                        class="
                            bg-[#3f4c66] shrink-0 h-5 w-5 rounded-full
                            grid place-items-center pointer
                            border border-transparent outline-none focus:border-cerise-strong hover:border-cerise-light relative
                        "
                        onClick={() => remove(passkey.id, i())}
                    >
                        <img class="w-3/5 h-3/5 invert" src="/public/x.svg" />
                    </button>
                </li>
            }</For>
        </ul>
        <Show
            when={adding()}
            fallback={<button
                onClick={() => setAdding(true)}
                class="
					bg-[#3f4c66] p-1.5 block rounded border text-center
					select-none border-transparent outline-none
					focus:border-cerise-strong hover:border-cerise-light
                "
            >Add passkey</button>}
        >
            <AddPasskey onAdded={passkey => {
                setPasskeys([...passkeys(), passkey]);
                setAdding(false);
            }} />
        </Show>
        <Show when={error() != ""}><p class="error">{error()}</p></Show>
    </section>;
}

function AddPasskey({ onAdded }: { onAdded: (passkey: Passkey) => void }) {
    let [name, setName] = createSignal("");
    let [errors, setErrors] = createSignal<string[]>([]);
    let cc: Promise<CredentialCreationOptions>;

    onMount(() => cc = (async () => {
        let res = await fetch("/passkey/add/begin", { method: "post" });
        let cc = await res.json();
        cc.publicKey.challenge = decodebase64url(cc.publicKey.challenge);
        cc.publicKey.user.id = decodebase64url(cc.publicKey.user.id);
        return cc;
    })());

    async function submit(event: Event) {
        event.preventDefault();
        setErrors([]);

        try {
            let cred: any;
            try {
                cred = await navigator.credentials.create(await cc);
            } catch {
                throw new Error("Missing permission or request was cancelled");
            }
            let res = await fetch("/passkey/add/finish", {
                method: "post",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    name: name(),
                    id: cred.id,
                    type: cred.type,
                    authenticatorAttachment: cred.authenticatorAttachment,
                    response: {
                        attestationObject: encodebase64url(cred.response.attestationObject),
                        clientDataJSON: encodebase64url(cred.response.clientDataJSON),
                    },
                }),
            });
            if (res.status != 200) throw new Error(await res.text());
            onAdded(passkeySchema.parse(await res.json()));
        } catch (err) {
            setErrors([
                ...errors(),
                (err instanceof Error)
                    ? err.message
                    : "Unkown error"
            ]);

        }
    }

    return <form onSubmit={submit}>
        <div class="flex gap-2">
            <input
                placeholder="passkey name"
                type="text"
                value={name()}
                onInput={e => setName(e.target.value)}
                autofocus
                class="
					border border-neutral-500 grow
					outline-none focus:border-cerise-strong hover:border-cerise-light
					bg-slate-800 p-1.5 rounded h-8
                "
            />
            <button class="
				bg-[#3f4c66] shrink-0 h-8 w-8 rounded-full
				grid place-items-center pointer
				border border-transparent outline-none focus:border-cerise-strong hover:border-cerise-light
            "><img class="w-3/5 h-3/5 invert" src="/public/check.svg" /></button>
        </div>
        <For each={errors()}>{error => <p class="error">{error}</p>}</For>
    </form>;
}

render(PasskeySettings, document.querySelector("#passkey-settings")!);
