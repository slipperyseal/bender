import bpy

for scene in bpy.data.scenes:
    scene.cycles.device = "GPU"
    scene.cycles.samples = {samples}
    scene.frame_start = {start}
    scene.frame_end = {end}
    scene.frame_step = 1
    scene.render.filepath = "{outpath}"
    scene.render.resolution_x = 3840
    scene.render.resolution_y = 1920
    scene.render.resolution_percentage = 100
    scene.render.use_motion_blur = True
    scene.render.motion_blur_shutter = 0.5
    scene.render.use_overwrite = True
    scene.render.use_border = False
    scene.render.image_settings.file_format = "OPEN_EXR"
    scene.render.image_settings.exr_codec = "ZIP" # NONE, PXR24, ZIP, PIZ, RLE, ZIPS, B44, B44A, DWAA, DWAB
    scene.render.image_settings.color_depth = "32"
    scene.render.image_settings.color_mode='RGB'

camera_name = "{camera}"
if camera_name:
    bpy.context.scene.camera = bpy.context.scene.objects.get(camera_name)
