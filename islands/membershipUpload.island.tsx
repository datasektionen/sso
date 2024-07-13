import { For, Show, createSignal } from "solid-js";
import { render } from "solid-js/web";

function MembershipUpload() {
    let fileInput!: HTMLInputElement;
    let [messages, setMessages] = createSignal<{ msg: string; error: boolean }[]>([]);
    let [progress, setProgress] = createSignal<number | null>(null);

    async function upload(event: Event) {
        event.preventDefault();

        setMessages([]);

        let res = await fetch("/admin/members/upload-sheet", {
            method: "post",
            headers: { "Content-Type": "application/octet-stream; charset=binary" },
            body: fileInput.files?.[0],
        });
        if (res.status != 200) {
            setMessages([{ msg: await res.text(), error: true }]);
            return
        }
        let events = new EventSource("/admin/members/upload-sheet");
        events.addEventListener("error", () => {
            events.close();
            setProgress(null);
            setMessages([...messages(), { msg: "Connection lost", error: true }]);
        });
        events.addEventListener("err", event => {
            setMessages([...messages(), { msg: event.data, error: true }]);
        });
        events.addEventListener("progress", event => {
            setProgress(parseFloat(event.data));
        });
        events.addEventListener("done", () => {
            setMessages([...messages(), { msg: "Upload finished", error: false }]);
            setProgress(null);
            events.close();
        });
    }

    return <form class="flex flex-col gap-2 p-2" onsubmit={upload}>
        <div class="flex gap-2">
            <input ref={fileInput} name="file" type="file" required class="
                bg-[#3f4c66] rounded p-1 grid place-items-center pointer w-full
                border border-transparent outline-none focus:border-cerise-strong hover:border-cerise-light relative
            " />
            <button class="
                bg-[#3f4c66] rounded p-1 grid place-items-center pointer
                border border-transparent outline-none focus:border-cerise-strong hover:border-cerise-light relative
            ">Upload</button>
        </div>
        <div class="w-full h-1" classList={{ "bg-white": progress() != null }}>
            <Show when={progress()}>{progress =>
                <div class="bg-cerise-regular h-full" style={{ width: (progress() * 100) + '%' }} />
            }</Show>
        </div>
        <For fallback each={messages()}>{message =>
            <p
                class="p-2 rounded"
                classList={{
                    "bg-red-600/50": message.error,
                    "bg-green-600/50": !message.error,
                }}
            >{message.msg}</p>
        }</For>
    </form>;
}

render(MembershipUpload, document.querySelector("#membership-upload")!);
