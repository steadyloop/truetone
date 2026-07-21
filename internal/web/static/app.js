// Truetone frontend behavior — two tiny bits, both no-op when their element is absent.

// Upload page: on submit, disable the button and show progress (processing
// happens in the background after that; the redirect lands back on the library).
const uploadForm = document.querySelector(".upload");
if (uploadForm) {
  uploadForm.addEventListener("submit", () => {
    const btn = uploadForm.querySelector("button[type=submit]");
    btn.textContent = "Transcoding…";
    btn.disabled = true;
  });
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
