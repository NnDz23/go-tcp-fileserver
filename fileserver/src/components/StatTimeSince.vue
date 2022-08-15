<template>
  <stat
    :header="header"
    :badge="badge"
    :prefix="prefix"
    :suffix="suffix"
  ></stat>
</template>
    
<script>
import Stat from "@/components/Stat.vue";
export default {
  components: { Stat },
  props: ["header", "prefix", "date", "suffix"],
  data() {
    return {
      interval: null,
      days: 0,
      hours: 0,
      minutes: 0,
      seconds: 0,
      badge: "",
      intervals: {
        second: 1000,
        minute: 1000 * 60,
        hour: 1000 * 60 * 60,
        day: 1000 * 60 * 60 * 24,
      },
    };
  },
  mounted() {
    this.interval = setInterval(() => {
      this.updateDiffs();
    }, this.intervals.minute);

    this.updateDiffs();
  },
  destroyed() {
    clearInterval(this.interval);
  },
  methods: {
    updateDiffs() {
      let diff = Math.abs(Date.now() - this.date.getTime());
      this.days = Math.floor(diff / this.intervals.day);
      diff -= this.days * this.intervals.day;
      this.hours = Math.floor(diff / this.intervals.hour);
      diff -= this.hours * this.intervals.hour;
      this.minutes = Math.floor(diff / this.intervals.minute);
      diff -= this.minutes * this.intervals.minute;
      this.seconds = Math.floor(diff / this.intervals.second);
      this.badge = `${this.days} days ${this.hours} hours ${this.minutes} minutes`;
    },
  },
};
</script>

<style>
</style>