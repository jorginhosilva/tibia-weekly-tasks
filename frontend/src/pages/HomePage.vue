<script setup>
import HeroSection from '@/components/home/HeroSection.vue'
import FeaturesSection from '@/components/home/FeaturesSection.vue'
import HowItWorksSection from '@/components/home/HowItWorksSection.vue'
import FaqSection from '@/components/home/FaqSection.vue'
import PartnerBanner from '@/components/layout/PartnerBanner.vue'
import { useSeo } from '@/composables/useSeo'
import { SITE } from '@/config'

useSeo({
  path: '/',
  schema: { '@context': 'https://schema.org', '@type': 'WebSite', name: SITE.name, url: SITE.url, inLanguage: 'pt-BR' },
})
</script>

<template>
  <HeroSection />
  <div class="px-4"><PartnerBanner /></div>
  <FeaturesSection />
  <HowItWorksSection />
  <FaqSection />
</template>
