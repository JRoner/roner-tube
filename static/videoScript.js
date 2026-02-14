document.addEventListener('DOMContentLoaded', () => {
    const el = document.querySelector('#dashPlayer')
    const id = el?.dataset.id
    if (!id) return
    const url = `/content/${id}/manifest.mpd`
    const player = dashjs.MediaPlayer().create()
    player.initialize(el, url, false)
})