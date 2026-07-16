const field = document.getElementById("nfc-id");
const button = document.getElementById("nfc-button");
const tooltip = document.getElementById("nfc-button-tooltip");
const error = document.getElementById("nfc-error")

if (!("NDEFReader" in window)) {
    button.classList.toggle('cursor-not-allowed');
    button.classList.toggle('opacity-50');
    button.disabled = true;
    tooltip.innerText = "Scan only supported in Chrome on Android"
}

async function startNFC() {
    try {
        button.classList.toggle('cursor-not-allowed');
        button.classList.toggle('opacity-50');
        button.disabled = true;

        const ndefReader = new NDEFReader();
        await ndefReader.scan();

        ndefReader.onreading = async ({ message, serialNumber }) => {
            // We use the serialNumber from the reader
            // Note: serialNumber availability can be experimental in some browsers
            // If serialNumber is not available, we'd need to parse an NDEF record

            field.value = serialNumber.toUpperCase();

            button.disabled = false;
            error.hidden = true;
            button.classList.toggle('cursor-not-allowed');
            button.classList.toggle('opacity-50');
        };

        ndefReader.onreadingerror = () => {
            error.innerText = "Failed to read card.";
            error.hidden = false;
            button.disabled = false;
            button.classList.toggle('cursor-not-allowed');
            button.classList.toggle('opacity-50');
        };
    } catch (error) {
        error.innerText = "Error: " + error.message;
        error.hidden = false;
        button.disabled = false;
    }
}
