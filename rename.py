import os

def rename_files_in_directory(directory):
    for root, _, files in os.walk(directory):
        for filename in files:
            if 'ibm-user-management-operator' in filename:
                new_filename = filename.replace('ibm-user-management-operator', 'ibm-user-management-operator')
                old_file = os.path.join(root, filename)
                new_file = os.path.join(root, new_filename)
                os.rename(old_file, new_file)
                print(f'Renamed: {old_file} to {new_file}')

# Specify the root directory containing the files
root_directory = '.'
rename_files_in_directory(root_directory)
