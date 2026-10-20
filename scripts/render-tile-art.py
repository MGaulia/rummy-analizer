r"""Render the pinned Printables model 130082 OBJ with Blender 4.3.2.

blender --background --threads 8 --python-exit-code 1 --python scripts/render-tile-art.py -- \
    --source /path/to/rummikub_joker.obj --output /path/to/tile-art

Source: Multicolor Rummikub tiles by KlaasYK, CC BY-SA 4.0.
https://www.printables.com/model/130082-multicolor-rummikub-tiles
The source mesh is unchanged; the script reorients it and assigns materials.
"""
from pathlib import Path
from collections import defaultdict
import argparse
import bpy
import hashlib
import json
import math
import sys

SOURCE_SHA256 = "cd92b514f2f435f1877b64e70e71c14716319e3bebdfa73c654f81cec905d1bc"
parser = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
parser.add_argument("--source", required=True, type=Path, help="Original rummikub_joker.obj")
parser.add_argument("--output", required=True, type=Path, help="Directory for PNG, JSON, and Blender scene")
arguments = sys.argv[sys.argv.index("--") + 1:] if "--" in sys.argv else []
args = parser.parse_args(arguments)
SOURCE = args.source.expanduser().resolve()
source_bytes = SOURCE.read_bytes()
actual_sha256 = hashlib.sha256(source_bytes).hexdigest()
if actual_sha256 != SOURCE_SHA256:
    raise SystemExit(f"Source SHA-256 mismatch: expected {SOURCE_SHA256}, got {actual_sha256}")
OUT = args.output.expanduser().resolve()
OUT.mkdir(parents=True, exist_ok=True)
COLORS = [("red", "c72c31"), ("blue", "245fbd"),
          ("black", "202124"), ("orange", "e58714")]
CELL_W, CELL_H = 192, 288

def rgb(hex_color):
    values = [int(hex_color[i:i+2], 16) / 255 for i in (0, 2, 4)]
    return tuple(v / 12.92 if v <= .04045 else ((v + .055) / 1.055) ** 2.4
                 for v in values) + (1,)

def material(name, color, roughness):
    mat = bpy.data.materials.new(name)
    mat.diffuse_color = rgb(color)
    mat.use_nodes = True
    node = mat.node_tree.nodes.get("Principled BSDF")
    node.inputs["Base Color"].default_value = mat.diffuse_color
    node.inputs["Roughness"].default_value = roughness
    return mat

def strip_render_metadata(path):
    """Keep PNG pixels intact while removing Blender's variable text metadata."""
    data = path.read_bytes()
    assert data[:8] == b"\x89PNG\r\n\x1a\n"
    chunks, offset = [data[:8]], 8
    while offset < len(data):
        length = int.from_bytes(data[offset:offset + 4], "big")
        kind = data[offset + 4:offset + 8]
        end = offset + length + 12
        assert end <= len(data)
        if kind not in {b"tEXt", b"zTXt", b"iTXt"}:
            chunks.append(data[offset:end])
        offset = end
    path.write_bytes(b"".join(chunks))

vertices, faces = [], defaultdict(lambda: defaultdict(list))
name = ""
for line in source_bytes.decode("utf-8").splitlines():
    if line.startswith("o "):
        name = line[2:]
    elif line.startswith("v "):
        vertices.append(tuple(map(float, line.split()[1:])))
    elif line.startswith("f "):
        face = [int(part.split("/")[0]) - 1 for part in line.split()[1:]]
        center_x = sum(vertices[i][0] for i in face) / len(face)
        center_y = sum(vertices[i][1] for i in face) / len(face)
        col = round(center_x / 25)
        source_row = round(-center_y / 35)
        assert 0 <= col < 7 and source_row in (0, 1)
        faces[name][source_row * 7 + col].append(face)

assert set(faces) == {"blocks_joker", "numbers_joker"}
assert all(set(by_tile) == set(range(14)) for by_tile in faces.values())
bpy.ops.object.select_all(action="SELECT")
bpy.ops.object.delete(use_global=False)
body = material("Ivory plastic", "f2eee2", .3)
meshes = {}
for part, by_tile in faces.items():
    for tile_index, part_faces in by_tile.items():
        cx, cy = (tile_index % 7) * 25, -(tile_index // 7) * 35
        indices = sorted({i for face in part_faces for i in face})
        remap = {old: new for new, old in enumerate(indices)}
        # The print face lies at -Z. Rotate 180 degrees about X to face up.
        coords = [(vertices[i][0] - cx, -(vertices[i][1] - cy), 2 - vertices[i][2])
                  for i in indices]
        mesh = bpy.data.meshes.new(f"source_{part}_{tile_index + 1}")
        mesh.from_pydata(coords, [], [[remap[i] for i in face] for face in part_faces])
        mesh.update()
        meshes[part, tile_index] = mesh

manifest = {
    "source": "https://www.printables.com/model/130082-multicolor-rummikub-tiles",
    "creator": "KlaasYK",
    "sourceObjSha256": SOURCE_SHA256,
    "license": "CC-BY-SA-4.0",
    "licenseUrl": "https://creativecommons.org/licenses/by-sa/4.0/",
    "modifications": "Reoriented original OBJ geometry, assigned ivory and four numeral colors, rendered atlas.",
    "image": "tiles-atlas.png", "width": 14 * CELL_W, "height": 4 * CELL_H,
    "cellWidth": CELL_W, "cellHeight": CELL_H, "sprites": []
}
for row, (color, hex_color) in enumerate(COLORS):
    ink = material(f"{color} numeral inlay", hex_color, .32)
    for tile_index in range(14):
        label = str(tile_index + 1) if tile_index < 13 else "joker"
        for part in ("blocks_joker", "numbers_joker"):
            mesh = meshes[part, tile_index].copy()
            mesh.materials.append(body if part == "blocks_joker" else ink)
            obj = bpy.data.objects.new(f"{color}_{label}_{part}", mesh)
            bpy.context.collection.objects.link(obj)
            obj.location = ((tile_index - 6.5) * 24, (1.5 - row) * 36, 0)
            obj.rotation_euler.x = math.radians(18)
            obj.rotation_euler.y = math.radians(-9)
        manifest["sprites"].append({
            "id": f"{color}-{label}", "color": color,
            "value": tile_index + 1 if tile_index < 13 else None,
            "joker": tile_index == 13,
            "x": tile_index * CELL_W, "y": row * CELL_H,
            "width": CELL_W, "height": CELL_H
        })

scene = bpy.context.scene
scene.render.engine = "CYCLES"
scene.cycles.device = "CPU"
scene.cycles.samples = 96
scene.cycles.seed = 130082
scene.cycles.use_animated_seed = False
scene.cycles.use_denoising = False
scene.render.resolution_x = manifest["width"]
scene.render.resolution_y = manifest["height"]
scene.render.resolution_percentage = 100
scene.render.film_transparent = True
scene.render.image_settings.file_format = "PNG"
scene.render.image_settings.color_mode = "RGBA"
scene.render.filepath = str(OUT / "tiles-atlas.png")
scene.view_settings.view_transform = "Standard"
scene.view_settings.look = "Medium High Contrast"
scene.world.use_nodes = True
scene.world.node_tree.nodes["Background"].inputs["Color"].default_value = (1, 1, 1, 1)
scene.world.node_tree.nodes["Background"].inputs["Strength"].default_value = .6

lamp = bpy.data.lights.new("Large softbox", "AREA")
lamp.energy = 250000
lamp.shape = "DISK"
lamp.size = 200
light_obj = bpy.data.objects.new("Large softbox", lamp)
bpy.context.collection.objects.link(light_obj)
light_obj.location = (-120, 100, 240)
light_obj.rotation_euler = (-light_obj.location).to_track_quat("-Z", "Y").to_euler()
camera_data = bpy.data.cameras.new("Orthographic atlas camera")
camera_data.type = "ORTHO"
camera_data.ortho_scale = 14 * 24
camera = bpy.data.objects.new("Orthographic atlas camera", camera_data)
bpy.context.collection.objects.link(camera)
camera.location = (0, 0, 500)
scene.camera = camera
(OUT / "tiles-atlas.json").write_text(json.dumps(manifest, indent=2) + "\n")
bpy.ops.wm.save_as_mainfile(filepath=str(OUT / "tiles-atlas.blend"))
bpy.ops.render.render(write_still=True)
strip_render_metadata(OUT / "tiles-atlas.png")
