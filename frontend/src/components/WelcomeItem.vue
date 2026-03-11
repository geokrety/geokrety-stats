<template>
  <div class="relative mt-8 flex gap-4 lg:mt-0 lg:py-2">
    <i
      class="flex h-8 w-8 items-center justify-center text-foreground lg:h-12 lg:w-12 lg:rounded-lg lg:border lg:border-border lg:bg-card"
    >
      <slot name="icon"></slot>
    </i>
    <div class="flex-1">
      <h3 class="mb-1 text-lg font-medium text-foreground">
        <slot name="heading"></slot>
      </h3>
      <slot></slot>
    </div>
  </div>
</template>
