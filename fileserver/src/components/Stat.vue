<template>
  <div class="card mt-2 mb-2">
    <div class="card-header">
      {{ header }}
    </div>
    <div class="card-body">
      <h5 class="card-title"> {{ prefix }} <span class="badge bg-secondary">{{ badge }}</span>
        {{ suffix }}</h5>
    </div>
  </div>
</template>

<script>
export default {
  name: "Stat",
  props: ["header","prefix","badge","suffix"],
};
</script>

<style>
</style>