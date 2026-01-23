import { showFlashMessage } from './flash.js';

// Copy to clipboard function
export function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
        showFlashMessage("URL copied to clipboard!", "success")
    }).catch(err => {
        console.error('Could not copy text: ', err);
    });
}

// Share or Copy function
export async function shareOrCopy(text) {
    if (navigator.share) {
        try {
            await navigator.share({
                title: 'Check out this link',
                url: text
            });
            // Optional: show success message for sharing? usually system handles it
        } catch (err) {
             // If user canceled or share failed, fall back to copy?
             // Usually if user cancels, we don't want to auto-copy.
             // If it's a real error (not AbortError), maybe copy.
             if (err.name !== 'AbortError') {
                console.error('Error sharing:', err);
                copyToClipboard(text);
             }
        }
    } else {
        // Fallback to clipboard
        copyToClipboard(text);
    }
}
