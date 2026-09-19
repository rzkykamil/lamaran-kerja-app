// Pemisah ribuan untuk input rupiah. Nilai yang dikirim tetap berupa angka
// karena server membuang semua karakter selain digit.
(function () {
  function format(angka) {
    return angka.replace(/\D/g, '').replace(/\B(?=(\d{3})+(?!\d))/g, '.');
  }

  function pasang(input) {
    input.value = format(input.value);
    input.addEventListener('input', function () {
      var posisiDariKanan = this.value.length - this.selectionStart;
      this.value = format(this.value);
      var posisi = Math.max(0, this.value.length - posisiDariKanan);
      this.setSelectionRange(posisi, posisi);
    });
  }

  document.querySelectorAll('.js-uang').forEach(pasang);
})();

// Modal konfirmasi kustom untuk form dengan atribut data-confirm, menggantikan
// confirm() bawaan browser yang tampilannya tidak bisa disesuaikan dengan tema.
(function () {
  var overlay = document.getElementById('modal-konfirmasi');
  if (!overlay) return;

  var pesan = document.getElementById('modal-konfirmasi-pesan');
  var tombolLanjut = overlay.querySelector('[data-modal-lanjut]');
  var tombolBatal = overlay.querySelector('[data-modal-batal]');
  var formTertunda = null;

  function tutup() {
    overlay.hidden = true;
    formTertunda = null;
  }

  document.addEventListener('submit', function (e) {
    var form = e.target;
    if (!(form instanceof HTMLFormElement) || !form.hasAttribute('data-confirm')) return;
    if (form === formTertunda) return; // sudah dikonfirmasi, biarkan lanjut submit

    e.preventDefault();
    formTertunda = form;
    pesan.textContent = form.getAttribute('data-confirm');
    overlay.hidden = false;
    tombolLanjut.focus();
  });

  tombolLanjut.addEventListener('click', function () {
    var form = formTertunda;
    tutup();
    if (form) form.requestSubmit();
  });

  tombolBatal.addEventListener('click', tutup);
  overlay.addEventListener('click', function (e) {
    if (e.target === overlay) tutup();
  });
  document.addEventListener('keydown', function (e) {
    if (e.key === 'Escape' && !overlay.hidden) tutup();
  });
})();
