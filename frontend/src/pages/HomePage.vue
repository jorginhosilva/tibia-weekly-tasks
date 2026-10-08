<script setup>
import HeroSection from '@/components/home/HeroSection.vue'
import FeaturesSection from '@/components/home/FeaturesSection.vue'
import HowItWorksSection from '@/components/home/HowItWorksSection.vue'
import FaqSection from '@/components/home/FaqSection.vue'
import { useSeo } from '@/composables/useSeo'
import { SITE } from '@/config'

useSeo({
  path: '/',
  schema: { '@context': 'https://schema.org', '@type': 'WebSite', name: SITE.name, url: SITE.url, inLanguage: 'pt-BR' },
})
</script>

<template>
  <HeroSection />
  <FeaturesSection />
  <HowItWorksSection />
  <FaqSection />
</template>
