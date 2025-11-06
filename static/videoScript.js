document.addEventListener('DOMContentLoaded', () => {
    const el = document.querySelector('#dashPlayer')
    const id = el?.dataset.id
    if (!id) return
    const url = `/content/${id}/manifest.mpd`
    const player = dashjs.MediaPlayer().create()
    player.initialize(el, url, false)

    // Nuke any attributes & inline styles that lock size >:)
    // const unlockSizing = () => {
    //     el.removeAttribute('width');
    //     el.removeAttribute('height');
    //     el.style.removeProperty('width');
    //     el.style.removeProperty('height');
    //     // In case dash.js or the browser re-adds them, assert our intent:
    //     el.style.setProperty('width', '50%');
    //     el.style.setProperty('height', '100%');
    //     el.style.maxWidth = '100%';
    // };
    //
    // unlockSizing();
    // el.addEventListener('loadedmetadata', unlockSizing);
    // el.addEventListener('loadeddata', unlockSizing);
})