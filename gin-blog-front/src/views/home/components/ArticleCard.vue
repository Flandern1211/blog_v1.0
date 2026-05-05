<script setup>
import { computed } from 'vue'
import dayjs from 'dayjs'

import { convertImgUrl } from '@/utils'

const props = defineProps({
  idx: Number,
  article: {},
})

const isRightClass = computed(() => props.idx % 2 === 0
  ? 'rounded-t-xl md:order-0 md:rounded-l-xl md:rounded-tr-0'
  : 'rounded-t-xl md:order-1 md:rounded-r-xl md:rounded-tl-0')
</script>

<template>
  <div class="group w-full flex flex-col animate-zoom-in animate-duration-700 items-center rounded-xl bg-white shadow-md transition-500 md:h-[200px] md:flex-row hover:shadow-2xl">
    <!-- 封面图 -->
    <div :class="isRightClass" class="h-[160px] w-full overflow-hidden md:h-full md:w-2/5">
      <RouterLink :to="`/article/${article.id}`">
        <img class="h-full w-full object-cover transition-600 hover:scale-105" :src="convertImgUrl(article.img)">
      </RouterLink>
    </div>
    <!-- 文章信息 -->
    <div class="my-3 w-9/10 md:w-3/5 space-y-2 md:px-6">
      <RouterLink :to="`/article/${article.id}`">
        <span class="text-lg font-bold leading-6 transition-300 group-hover:text-violet line-clamp-1">
          {{ article.title }}
        </span>
      </RouterLink>
      <div class="flex flex-wrap items-center text-xs color-[#858585]">
        <span v-if="article.is_top === 1" class="flex items-center color-[#ff7242] mr-1.5">
          <span class="i-carbon:align-vertical-top mr-0.5" /> 置顶
        </span>
        <span class="flex items-center mr-1.5">
          <span class="i-mdi-calendar-month-outline mr-0.5" /> {{ dayjs(article.created_at).format('YYYY-MM-DD') }}
        </span>
        <span class="mr-1.5">|</span>
        <RouterLink :to="`/categories/${article.category_id}?name=${article.category?.name}`" class="flex items-center mr-1.5">
          <span class="i-mdi-inbox-full mr-0.5" /> {{ article.category?.name }}
        </RouterLink>
        <span class="mr-1.5">|</span>
        <div class="flex gap-1">
          <RouterLink v-for="tag in article.tags" :key="tag.id" :to="`/tags/${tag.id}?name=${tag.name}`" class="flex items-center">
            <span class="i-mdi-tag-multiple mr-0.5" /> {{ tag.name }}
          </RouterLink>
        </div>
      </div>
      <p class="text-xs color-[#666] leading-5 line-clamp-3">
        {{ article.content }}
      </p>
    </div>
  </div>
</template>
