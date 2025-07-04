<script setup lang="ts">
import {ref, watch} from "vue";
import {debounce} from "lodash-es";
import {propTypes} from "@/utils/propTypes";
import {ElIcon, ElImage} from "element-plus";
import {Picture as IconPicture} from '@element-plus/icons-vue'

// ---------------------------------
// common
// ---------------------------------
const props = defineProps({
  image: propTypes.string.def(''),
  background: propTypes.bool.def(false),
})

// ---------------------------------
// component methods
// ---------------------------------

const Image = ref<string>()
const style = ref({})

const update = debounce(async () => {
  if (props.image) {
    Image.value = props.image
    style.value = {
      "background": `url(${props.image})`,
      'width': '100%',
      'height': '100%'
    }
  } else {
    Image.value = null
    style.value = {}
  }
}, 100)

watch(
  () => [props.image, props.background],
  (val) => {
    update()
  },
  {
    deep: true,
    immediate: true
  }
)

</script>

<template>
  <ElImage v-if="!props.background && image" :src="Image">
    <template #error>
      <div class="image-slot">
        <ElIcon>
          <icon-picture/>
        </ElIcon>
      </div>
    </template>
  </ElImage>
  <div v-else :style="style"></div>
</template>

<style lang="less" scoped>
.el-image__error,
.el-image__placeholder,
.el-image__inner {
  height: auto;
}

.el-image.item-element {
  overflow: visible;
}
</style>
