require 'json'

CONFIG_json_content = {
	command: 'echo "Hello: ${T_KEY}!"',
	environments: [
		key: "T_KEY",
		value: "test value, with equal = sign, for test"
	]
}

# CONFIG_json_content = {
# 	command: 'env'
# }

# CONFIG_json_content = {
# 	command: 'echo "$HOME"',
# 	environments: [
# 		key: "HOME",
# 		value: "my-test-home"
# 	]
# }

# CONFIG_json_content = {
# 	command: 'ls',
# 	working_directory: '/Users'
# }

# puts JSON.dump( JSON.generate(CONFIG_json_content) )
puts JSON.generate(CONFIG_json_content)
