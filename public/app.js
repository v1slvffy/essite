const form = document.getElementById('screenshot-form');
const urlInput = document.getElementById('url-input');
const fullpageInput = document.getElementById('fullpage-input');
const btnSubmit = document.getElementById('btn-submit');
const statusBox = document.getElementById('status-box');
const resultSection = document.getElementById('result-section');
const previewImage = document.getElementById('preview-image');
const previewMeta = document.getElementById('preview-meta');
const btnDownload = document.getElementById('btn-download');
const btnCopyUrl = document.getElementById('btn-copy-url');

let currentImageUrl = '';

form.addEventListener('submit', async (e) => {
  e.preventDefault();

  const url = urlInput.value.trim();
  if (!url) return;

  const device = document.querySelector('input[name="device"]:checked')?.value || 'desktop';
  const fullPage = fullpageInput.checked;

  btnSubmit.disabled = true;
  btnSubmit.textContent = 'Sedang Mengambil Screenshot...';
  statusBox.className = 'status-box loading';
  statusBox.textContent = `Memuat tangkapan layar untuk ${url} (${device})... Mohon tunggu beberapa detik.`;
  statusBox.style.display = 'block';
  resultSection.style.display = 'none';

  try {
    const params = new URLSearchParams({
      url,
      device,
      fullPage: String(fullPage),
    });

    const res = await fetch(`/api/capture?${params.toString()}`);
    const rawText = await res.text();
    let data;

    try {
      data = JSON.parse(rawText);
    } catch {
      throw new Error(`Server (${res.status}): ${rawText.slice(0, 150) || res.statusText || 'Respon bukan format JSON'}`);
    }

    if (!res.ok || !data.status) {
      throw new Error(data.error || `Gagal mengambil screenshot (${res.status}).`);
    }

    const result = data.result;
    currentImageUrl = result.imageUrl;

    previewImage.src = currentImageUrl;
    previewMeta.textContent = `Target: ${result.targetUrl} | Device: ${result.device} | Dimensi: ${result.dimensions?.width || '-'}x${result.dimensions?.height || '-'}px | Ukuran: ${result.fileSize || '-'}`;

    btnDownload.href = currentImageUrl;
    btnDownload.download = `screenshot-${result.device}.png`;

    statusBox.style.display = 'none';
    resultSection.style.display = 'block';
  } catch (err) {
    statusBox.className = 'status-box error';
    statusBox.textContent = `Error: ${err.message || 'Terjadi kesalahan saat memproses screenshot.'}`;
    statusBox.style.display = 'block';
  } finally {
    btnSubmit.disabled = false;
    btnSubmit.textContent = 'Ambil Screenshot';
  }
});

btnCopyUrl.addEventListener('click', async () => {
  if (!currentImageUrl) return;
  try {
    await navigator.clipboard.writeText(currentImageUrl);
    const oldText = btnCopyUrl.textContent;
    btnCopyUrl.textContent = 'Tersalin!';
    setTimeout(() => {
      btnCopyUrl.textContent = oldText;
    }, 2000);
  } catch {
    alert('Gagal menyalin link.');
  }
});
