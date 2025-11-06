const toggle = document.querySelector('.nav-toggle');
const links  = document.querySelector('.nav-links');

toggle.addEventListener('click', () => {
    links.checkVisibility()
    links.classList.toggle('open');
});

const form = document.querySelector('.video-upload');
const buttonsWrap  = document.querySelector('.upload-buttons');
const uploadingWrap = document.querySelector('.uploading');
const uploadBtn = document.getElementById('submit-button');

if (form && buttonsWrap && uploadingWrap && uploadBtn) {
    form.addEventListener('submit', (e) => {
        // swap UI
        buttonsWrap.style.display = 'none';
        uploadingWrap.style.display = 'flex';
        uploadBtn.disabled = true;
    });
}
