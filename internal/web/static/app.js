// Truetone frontend behavior — two tiny bits, both no-op when their element is absent.

// Library: nav Upload button opens the hidden file input, then auto-submits and
// flips to "Transcoding…" (phase-0 transcode is synchronous, so the page waits).
const file = document.getElementById("file");
const btn = document.getElementById("uploadBtn");
if (file && btn) {
  btn.onclick = () => file.click();
  file.onchange = () => {
    if (!file.files.length) return;
    btn.textContent = "Transcoding…";
    btn.disabled = true;
    file.form.submit();
  };
}

// Watch: attach HLS to the <video>. Safari plays it natively; everyone else uses hls.js.
const video = document.getElementById("video");
if (video) {
  const src = video.dataset.src;
  if (video.canPlayType("application/vnd.apple.mpegurl")) {
    video.src = src;
  } else if (window.Hls && Hls.isSupported()) {
    const hls = new Hls();
    hls.loadSource(src);
    hls.attachMedia(video);
  }
}
