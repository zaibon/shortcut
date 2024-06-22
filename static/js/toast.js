document.body.addEventListener("makeToast", onMakeToast);

/**
 * Presents a toast notification when the `makeToast` event is triggered
 * @param e {{detail: {level: string, title: string, message: string}[]}}
 */
function onMakeToast(e) {
    console.log(e);
    const level = e.detail.level;
    const message = e.detail.message;
    const title = e.detail.title;
    const customWrapper = e.detail.customWrapper;

    new Notify({
        title: title,
        text: message,
        status: level,
        customWrapper: customWrapper ? customWrapper : null,

        showIcon: true,
        showCloseButton: true,
        autoclose: false,
        // autotimeout: 3000,
        gap: 20,
        distance: 20,
        type: 'outline',
        position: 'right bottom'
    })
}
