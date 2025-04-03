// Copy to clipboard function
function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
        showFlashMessage("URL copied to clipboard!", "success")
    }).catch(err => {
        console.error('Could not copy text: ', err);
    });
}