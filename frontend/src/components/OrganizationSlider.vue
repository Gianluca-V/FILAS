<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';

import { resolveAssetPath } from '../utils/assets.js';

// Reactive reimplementation of the legacy clip-path carousel
// (js/organization.mjs's populateSlider): reactive `activeIndex` replaces
// direct DOM class toggling, and the setInterval-based autoplay
// pauses/resumes on hover exactly like the original.
const props = defineProps({
  organizations: {
    type: Array,
    required: true,
  },
});

const activeIndex = ref(0);
let autoplayTimer = null;

const activeCount = computed(() => props.organizations.length);

function next() {
  if (activeCount.value === 0) return;
  activeIndex.value = (activeIndex.value + 1) % activeCount.value;
}

function previous() {
  if (activeCount.value === 0) return;
  activeIndex.value = (activeIndex.value - 1 + activeCount.value) % activeCount.value;
}

function startAutoplay() {
  stopAutoplay();
  autoplayTimer = setInterval(next, 10000);
}

function stopAutoplay() {
  if (autoplayTimer) {
    clearInterval(autoplayTimer);
    autoplayTimer = null;
  }
}

onMounted(startAutoplay);
onBeforeUnmount(stopAutoplay);
</script>

<template>
  <div class="slider" @mouseover="stopAutoplay" @mouseout="startAutoplay">
    <div
      v-for="(organization, index) in organizations"
      :key="organization.ID"
      class="slide"
      :class="{ active: index === activeIndex }"
    >
      <div class="img-abogado" :style="{ backgroundImage: `url('${resolveAssetPath(organization.Image)}')` }" />
      <div class="info">
        <h2>{{ organization.Title }}</h2>
        <p>{{ organization.Description }}</p>
      </div>
    </div>

    <div v-if="activeCount > 1" class="navigation">
      <button type="button" class="prev-btn" aria-label="Anterior" @click="previous">&lt;</button>
      <button type="button" class="next-btn" aria-label="Siguiente" @click="next">&gt;</button>
    </div>
  </div>
</template>

<style scoped lang="scss">
.slider {
  position: relative;
  background-color: transparent;
  width: 75%;
  min-height: 500px;
  margin: auto;
  overflow: hidden;
  z-index: 900;
  border-radius: 20px;
}

.slide {
  position: absolute;
  width: 100%;
  height: 100%;
  clip-path: circle(0% at 0 50%);
  display: flex;

  &.active {
    clip-path: circle(150% at 0 50%);
    transition: 1s;
  }
}

.img-abogado {
  width: 50%;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
}

.info {
  color: #222;
  width: 50%;
  padding: 25px;
  background-color: $color-background;
  @include flex-column(center, flex-start);

  h2 {
    font-size: 2em;
    color: #152744;
  }

  p {
    font-size: 1em;
    color: black;
    line-height: 21px;
    font-weight: 400;
    text-align: justify;
  }
}

.navigation {
  height: 500px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  opacity: 0;
  transition: opacity 0.5s ease;
  position: relative;
  z-index: 999;
}

.slider:hover .navigation {
  opacity: 1;
}

.prev-btn,
.next-btn {
  font-size: 1.5rem;
  color: #222;
  background: rgba(255, 255, 255, 0.4);
  padding: 10px;
  cursor: pointer;
  box-shadow: 0 0 2px black;
}

.prev-btn {
  border-top-right-radius: 3px;
  border-bottom-right-radius: 3px;
}

.next-btn {
  border-top-left-radius: 3px;
  border-bottom-left-radius: 3px;
}

@include below-tablet {
  .info {
    position: relative;
    width: 80%;
    margin-inline: auto;
  }
}

@media (max-width: 650px) {
  .slider {
    height: 100%;
  }

  .slide {
    flex-direction: column;
  }

  .img-abogado {
    background-position: center;
    height: 35%;
    width: 100%;
  }

  .info {
    height: 65%;
    width: 100%;
  }
}
</style>
