var reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

// ---- Big photo view ----

var lightbox = document.querySelector('.lightbox');
var lightboxImg = lightbox && lightbox.querySelector('.lightbox-img');
var lightboxPhotos = []; // the <img> elements of the bread being viewed
var lightboxIndex = 0;

function showLightboxPhoto(index) {
  var count = lightboxPhotos.length;
  lightboxIndex = (index + count) % count; // wraps around
  lightboxImg.src = lightboxPhotos[lightboxIndex].src;
  lightboxImg.alt = lightboxPhotos[lightboxIndex].alt;
}

function openLightbox(photos, index) {
  lightboxPhotos = photos;
  lightbox.classList.toggle('lightbox-single', photos.length < 2);
  showLightboxPhoto(index);
  lightbox.showModal(); // Esc closes it
}

if (lightbox) {
  lightbox.querySelector('.lightbox-close').addEventListener('click', function () {
    lightbox.close();
  });
  lightbox.querySelector('.lightbox-prev').addEventListener('click', function () {
    showLightboxPhoto(lightboxIndex - 1);
  });
  lightbox.querySelector('.lightbox-next').addEventListener('click', function () {
    showLightboxPhoto(lightboxIndex + 1);
  });

  // Clicking the dark area around the photo closes it.
  lightbox.addEventListener('click', function (event) {
    if (event.target === lightbox) {
      lightbox.close();
    }
  });

  // Left/right arrow keys change the photo.
  lightbox.addEventListener('keydown', function (event) {
    if (lightboxPhotos.length < 2) return;
    if (event.key === 'ArrowLeft') showLightboxPhoto(lightboxIndex - 1);
    if (event.key === 'ArrowRight') showLightboxPhoto(lightboxIndex + 1);
  });
}

// ---- Each bread box ----

document.querySelectorAll('.bread-box').forEach(function (box) {
  var track = box.querySelector('.slider-track');
  if (!track) {
    return; // "Foto próximamente" placeholder: no photos
  }

  var photos = Array.prototype.slice.call(track.querySelectorAll('img'));

  if (lightbox) {
    photos.forEach(function (img, i) {
      img.addEventListener('click', function () { openLightbox(photos, i); });
    });
  }

  var prev = box.querySelector('.slider-prev');
  var next = box.querySelector('.slider-next');
  var count = photos.length;

  if (!prev || count < 2) {
    return; // only one photo: nothing to slide
  }

  function currentIndex() {
    return Math.round(track.scrollLeft / track.clientWidth);
  }

  // Wraps around: "next" on the last photo goes back to the first.
  function goTo(index) {
    index = (index + count) % count;
    track.scrollTo({
      left: index * track.clientWidth,
      behavior: reduceMotion ? 'auto' : 'smooth'
    });
  }

  prev.addEventListener('click', function () { goTo(currentIndex() - 1); });
  next.addEventListener('click', function () { goTo(currentIndex() + 1); });
});

// ---- Menu page ----

// Clicking a menu photo opens all of that item's photos in the big view.
document.querySelectorAll('.menu-item-photos').forEach(function (group) {
  var photos = Array.prototype.slice.call(group.querySelectorAll('img'));
  if (lightbox && photos.length) {
    photos[0].addEventListener('click', function () { openLightbox(photos, 0); });
  }
});
