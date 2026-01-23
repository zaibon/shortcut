// QR Code Component for Shortcut
export function registerQrCodeModal(Alpine) {
    Alpine.data("qrCodeModal", (initialUrl = "") => ({

        open: false,
        url: initialUrl,
        qrSize: "medium",
        qrColor: "#000000",

        // Computed property for QR code URL
        get qrCodeUrl() {
            const size = this.qrSize === "small" ? "150x150" : this.qrSize === "medium" ? "200x200" : "300x300"
            const color = this.qrColor.replace("#", "")
            return `https://api.qrserver.com/v1/create-qr-code/?size=${size}&data=${encodeURIComponent(this.url)}&color=${color}`
        },

        // Method to show the modal with a specific URL
        showModal(url) {
            this.url = url
            this.open = true
        },

        // Method to download the QR code
        downloadQRCode() {
            // Create a temporary link element
            const link = document.createElement("a")
            link.href = this.qrCodeUrl
            link.download = `qrcode-${this.url.replace(/[^a-zA-Z0-9]/g, "-")}.png`
            link.target = "_blank"
            document.body.appendChild(link)
            link.click()
            document.body.removeChild(link)
        },
    }))
}
