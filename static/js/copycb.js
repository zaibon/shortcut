const copyContent = async (copyContent, copyIcon, doneIcon) => {
    let copyIconElement = document.getElementById(copyIcon);
    let doneIconElement = document.getElementById(doneIcon);
    let text = document.getElementById(copyContent).innerHTML;
    try {
        console.log('copying', text);
        await navigator.clipboard.writeText(text);

        copyIconElement.classList.remove("opacity-100");
        copyIconElement.classList.add("opacity-0");

        doneIconElement.classList.remove("opacity-0");
        doneIconElement.classList.add("opacity-100");

        setTimeout(() => {
            copyIconElement.classList.remove("opacity-0");
            copyIconElement.classList.add("opacity-100");

            doneIconElement.classList.remove("opacity-100");
            doneIconElement.classList.add("opacity-0");
        }, 1000);
    } catch (err) {
        console.log('failed to copy!', err);
    }
}