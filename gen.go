package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func downloadFile(filePath string, url string) error {
	outFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}

	_, err = io.Copy(outFile, resp.Body)
	return err
}

func downloadSokol() {
	dir := "sokol"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			fmt.Printf("Failed to create directory: %v\n", err)
			return
		}
	}

	files := []string{
		"sokol_app.h",
		"sokol_gfx.h",
		"sokol_audio.h",
		"sokol_time.h",
		"sokol_args.h",
		"sokol_fetch.h",
		"sokol_log.h",
		"sokol_glue.h",
	}

	baseUrl := "https://raw.githubusercontent.com/floooh/sokol/refs/heads/master/"

	for _, file := range files {
		url := baseUrl + file
		destPath := filepath.Join(dir, file)
		if err := downloadFile(destPath, url); err != nil {
			fmt.Printf("Error downloading %s: %v\n", file, err)
		} else {
			fmt.Printf("Successfully downloaded %s\n", file)
		}
	}
}

func main() {
	mainfile, err := os.Create("main.c")
	if err != nil {
		fmt.Println("Error creating main file")
		return
	}

	defer mainfile.Close()

	var input int
	fmt.Println("Enter target platform 1-6: ")
	fmt.Println("1: GLCORE")
	fmt.Println("2: GLES3")
	fmt.Println("3: D3D11")
	fmt.Println("4: METAL")
	fmt.Println("5: WGPU")
	fmt.Println("6: NOAPI")
	_, err = fmt.Scanln(&input)
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}

	var api string
	switch input {
	case 1:
		api = "SOKOL_GLCORE"
	case 2:
		api = "SOKOL_GLES3"
	case 3:
		api = "SOKOL_D3D11"
	case 4:
		api = "SOKOL_METAL"
	case 5:
		api = "SOKOL_WGPU"
	case 6:
		api = "SOKOL_NOAPI"
	}
	_, err = mainfile.WriteString(
		"#define SOKOL_IMPL\n" +
			"#define " + api + "\n" +
			"#include \"sokol/sokol_gfx.h\"\n" +
			"#include \"sokol/sokol_app.h\"\n" +
			"#include \"sokol/sokol_log.h\"\n" +
			"#include \"sokol/sokol_glue.h\"\n" +
			"#include \"sokol/sokol_audio.h\"\n" +
			"\n" +
			"static void init(void) {\n" +
			"\tsg_desc desc = { };\n" +
			"\tdesc.environment = sglue_environment();\n" +
			"\tdesc.logger.func = slog_func;\n" +
			"\tsg_setup(&desc);\n" +
			"}\n" +
			"\n" +
			"static void event(const sapp_event* e) {\n" +
			"}\n" +
			"\n" +
			"static void frame(void) {\n" +
			"\tchar window_title[64];\n" +
			"\tsnprintf(window_title, sizeof(window_title), \"Standin Title\", (int) sapp_frame_count(), sapp_frame_duration()*1000.0);\n" +
			"\tsapp_set_window_title(window_title);\n" +
			"\tsg_pass_action pass_action = {\n" +
			"\t\t.colors[0] = {\n" +
			"\t\t\t.load_action = SG_LOADACTION_CLEAR,\n" +
			"\t\t\t.clear_value = { 1.0f, 0.0f, 0.0f, 1.0f }\n" +
			"\t\t}\n" +
			"\t};\n" +
			"\tsg_begin_pass(&(sg_pass){ .action = pass_action, .swapchain = sglue_swapchain()});\n" +
			"\tsg_end_pass();\n" +
			"\tsg_commit();\n" +
			"}\n" +
			"\n" +
			"static void cleanup(void) {\n" +
			"\tsg_shutdown();\n" +
			"}\n" +
			"\n" +
			"sapp_desc sokol_main(int argc, char* argv[]) {\n" +
			"\t(void)argc;\n" +
			"\t(void)argv;\n" +
			"\tsapp_desc desc = { };\n" +
			"\tdesc.init_cb = init;\n" +
			"\tdesc.frame_cb = frame;\n" +
			"\tdesc.event_cb = event;\n" +
			"\tdesc.cleanup_cb = cleanup;\n" +
			"\tdesc.width = 1200;\n" +
			"\tdesc.height = 800;\n" +
			"\tdesc.window_title = \"Standin Title\";\n" +
			"\tdesc.icon.sokol_default = true;\n" +
			"\tdesc.logger.func = slog_func;\n" +
			"\treturn desc;\n" +
			"}\n",
	)
	if err != nil {
		fmt.Println("Error writing to file")
		return
	}

	makefile, err := os.Create("Makefile")
	if err != nil {
		fmt.Println("Error creating makefile")
		return
	}

	_, err = makefile.WriteString(
		"EMCC = emcc\n" +
			"CFLAGS = -Wall -O2 -std=c99\n" +
			"INCLUDES = -Isokol\n" +
			"SOURCES_C = main.c\n" +
			"\n\n" +
			"LIBS = -lm -lopengl32 -lgdi32\n" +
			"\n\n" +
			"WASM_OBJS = $(SOURCES_C:.c=.wasm.o)\n" +
			"\n" +
			"wasm: CC=$(EMCC)\n" +
			"wasm: CFLAGS += -D" + api + " -s WASM=1 -s ALLOW_MEMORY_GROWTH=1 -sASSERTIONS -s ASYNCIFY\n" +
			"wasm: LIBS=-sUSE_WEBGL2 -s ALLOW_MEMORY_GROWTH=1\n" +
			"wasm: OUTPUT = sokol_out.html\n" +
			"wasm: SHELL_FILE = sokol.html\n" +
			"\n" +
			"wasm: $(WASM_OBJS)\n" +
			"\t$(EMCC) $(CFLAGS) $(INCLUDES) $^ -o sokol.js --shell-file $(SHELL_FILE) $(LIBS)\n" +
			"\n\n" +
			"%.0: %.c\n" +
			"\t$(CC) $(CFLAGS) $(INCLUDES) -c $< -o $@\n" +
			"\n" +
			"%.wasm.o: %.c\n" +
			"\t$(EMCC) $(CFLAGS) $(INCLUDES) -c $< -o $@\n",
	)
	if err != nil {
		fmt.Println("Error writing to file")
		return
	}

	shellhtml, err := os.Create("sokol.html")
	if err != nil {
		fmt.Println("Error creating html shell")
		return
	}

	_, err = shellhtml.WriteString(
		"<!DOCTYPE html>\n" +
			"<html lang=\"en\">\n" +
			"<head>\n" +
			"\t<link rel=\"stylesheet\" href=\"style.css\">\n" +
			"\t<meta charset=\"UTF-8\">\n" +
			"\t<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n" +
			"\t<title>Sokol</title>\n" +
			"</head>\n" +
			"<body>\n" +
			"\t<canvas id=\"canvas\"></canvas>\n" +
			"\t<script src=\"sokol.js\"></script>\n" +
			"</body>\n" +
			"</html>\n",
	)
	if err != nil {
		fmt.Println("Error writing to html shell")
		return
	}

	cssfile, err := os.Create("style.css")
	if err != nil {
		fmt.Println("Error creating css file")
		return
	}

	_, err = cssfile.WriteString(
		"#canvas {\n" +
			"\theight: 100%;\n" +
			"\twidth: 100%;\n" +
			"}\n",
	)
	if err != nil {
		fmt.Println("Error writing to css")
		return
	}

	downloadSokol()
}
