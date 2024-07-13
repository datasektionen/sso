import { For, Index, Show, createEffect, createSignal, onMount } from "solid-js";
import { render } from "solid-js/web";
import { z } from "zod";

let clientSchema = z.object({
    id: z.string(),
    redirect_uris: z.string().array(),
});

type Client = z.infer<typeof clientSchema>;

function OidcClients() {
    let [clients, setClients] = createSignal<Client[]>([]);
    let [error, setError] = createSignal<string>("");
    let [loading, setLoading] = createSignal(true);
    let [adding, setAdding] = createSignal(false);
    let [secret, setSecret] = createSignal<string | null>(null);

    onMount(async () => {
        let res = await fetch("/admin/list-oidc-clients");
        if (res.status != 200) {
            setError(await res.text());
            return;
        }
        let body = await res.json();
        setClients(z.array(clientSchema).parse(body));
        setLoading(false);
    });

    async function remove(id: string, index: number) {
        let res = await fetch("/admin/oidc-clients/" + id, {
            method: "delete",
            headers: { "Content-Type": "application/json" },
        });
        if (res.status != 200) return setError(await res.text());
        let c = clients();
        setClients([...c.slice(0, index), ...c.slice(index + 1)]);
    }

    return <section class="flex flex-col">
        <h2 class="text-lg">OIDC Clients:</h2>
        <ul>
            <Show when={loading()}>...</Show>
            <For each={clients()}>{(client, i) =>
                <li class="p-2">
                    <div class="flex gap-2 items-center">
                        <span>{client.id}</span>
                        <button
                            class="
                                bg-[#3f4c66] shrink-0 h-5 w-5 rounded-full
                                grid place-items-center pointer
                                border border-transparent outline-none focus:border-cerise-strong hover:border-cerise-light relative
                            "
                            onClick={() => remove(client.id, i())}
                        >
                            <img class="w-3/5 h-3/5 invert" src="/public/x.svg" />
                        </button>
                    </div>
                    <ul class="pl-3"><For each={client.redirect_uris}>{uri =>
                        <li>{uri}</li>
                    }</For></ul>
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
            >New client</button>}
        >
            <AddClient onAdded={(client, secret) => {
                setClients([...clients(), client]);
                setAdding(false);
                setSecret(secret);
            }} />
        </Show>
        <Show when={secret()}>{secret => <p>Secret: <code>{secret()}</code></p>}</Show>
        <Show when={error() != ""}><p class="error">{error()}</p></Show>
    </section>;
}

function AddClient({ onAdded }: { onAdded: (client: Client, secret: string) => void }) {
    let [redirectUris, setRedirectUris] = createSignal([""]);
    let [error, setError] = createSignal<string | null>(null);

    function addUri(event: Event) {
        event.preventDefault();
        setRedirectUris([...redirectUris(), ""]);
    }

    function setUri(index: number, value: string) {
        setRedirectUris([...redirectUris().slice(0, index), value, ...redirectUris().slice(index + 1)]);
    }

    function removeUri(index: number) {
        setRedirectUris([...redirectUris().slice(0, index), ...redirectUris().slice(index + 1)]);
    }

    createEffect(() => console.log(redirectUris()));

    async function submit() {
        setError("");

        let res = await fetch("/admin/oidc-clients", {
            method: "post",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                redirect_uris: redirectUris(),
            }),
        });
        if (res.status != 200) return setError(await res.text());
        let json = await res.json();
        let secret = z.string().parse(json.secret);
        onAdded(clientSchema.parse({
            id: json.id,
            redirect_uris: redirectUris(),
        }), secret);
    }

    return <>
        <div class="flex gap-2 flex-col">
            <Index each={redirectUris()}>{(uri, i) => <div class="flex items-center gap-2">
                <input
                    type="text"
                    value={uri()}
                    onInput={event => setUri(i, event.target.value)}
                    class="
                        border border-neutral-500 grow
                        outline-none focus:border-cerise-strong hover:border-cerise-light
                        bg-slate-800 p-1.5 rounded h-8
                    "
                />
                <button
                    class="
                        bg-[#3f4c66] shrink-0 h-6 w-6 rounded-full
                        grid place-items-center pointer
                        border border-transparent outline-none focus:border-cerise-strong hover:border-cerise-light relative
                    "
                    onClick={() => removeUri(i)}
                >
                    <img class="w-3/5 h-3/5 invert" src="/public/x.svg" />
                </button>
            </div>}</Index>
            <div class="flex gap-2">
                <button onclick={addUri} class="
                    bg-[#3f4c66] p-1.5 block rounded border text-center grow
                    select-none border-transparent outline-none
                    focus:border-cerise-strong hover:border-cerise-light
                ">Additional redirect uri</button>
                <button onclick={submit} class="
                    bg-[#3f4c66] p-1.5 block rounded border text-center grow
                    select-none border-transparent outline-none
                    focus:border-cerise-strong hover:border-cerise-light
                ">Submit</button>
            </div>
        </div>
        <Show when={error()}>{error => <p class="error">{error()}</p>}</Show>
    </>;
}
render(OidcClients, document.querySelector("#oidc-clients")!);
